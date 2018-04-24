package controller

import (
	"bytes"
	"context"
	"fmt"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/AlecAivazis/survey"
	arg "github.com/alexflint/go-arg"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/ktr0731/evans/adapter/gateway"
	"github.com/ktr0731/evans/adapter/parser"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/evans/usecase/port"
	updater "github.com/ktr0731/go-updater"
	"github.com/ktr0731/go-updater/brew"
	"github.com/ktr0731/go-updater/github"
	isatty "github.com/mattn/go-isatty"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/pkg/errors"

	"io"
	"os"
)

var (
	ErrProtoFileRequired = errors.New("least one proto file required")
)

type Options struct {
	Proto []string `arg:"positional,help:.proto files"`

	Interactive bool     `arg:"-i,help:use interactive mode"`
	EditConfig  bool     `arg:"-e,help:edit config file by $EDITOR"`
	Host        string   `arg:"help:gRPC host"`
	Port        string   `arg:"-p,help:gRPC port"`
	Package     string   `arg:"help:default package"`
	Service     string   `arg:"help:default service. evans parse package from this if --package is nothing."`
	Call        string   `arg:"-c,help:call specified RPC"`
	File        string   `arg:"-f,help:the script file which will be executed (used only command-line mode)"`
	Path        []string `arg:"separate,help:proto file path"`
	Header      []string `arg:"separate,help:headers set to each requests"`

	name    string `arg:"-"`
	version string `arg:"-"`
}

func (o *Options) Version() string {
	return fmt.Sprintf("%s %s", o.name, o.version)
}

type CLI struct {
	ui     ui
	config *config.Config

	parser  *arg.Parser
	options *Options

	cache *meta.Meta
}

func NewCLI(name, version string) *CLI {
	return &CLI{
		ui: newUI(),
		options: &Options{
			Port:    "50051",
			name:    name,
			version: version,
		},
		config: config.Get(),
		cache:  meta.Get(),
	}
}

func (c *CLI) Error(err error) {
	c.ui.ErrPrintln(err.Error())
}

func (c *CLI) Help() {
	c.parser.WriteHelp(c.ui.Writer())
}

func (c *CLI) Usage() {
	c.parser.WriteUsage(c.ui.Writer())
}

func (c *CLI) Run(args []string) int {
	c.parser = arg.MustParse(c.options)

	if c.options.EditConfig {
		if err := config.Edit(); err != nil {
			c.Error(err)
			return 1
		}
		return 0
	}

	if err := checkPrecondition(c.config, c.options); err != nil {
		c.Error(err)
		return 1
	}

	env, err := setupEnv(c.config, c.options)
	if err == ErrProtoFileRequired {
		c.Usage()
	}
	if err != nil {
		c.Error(err)
		return 1
	}

	grpcAdapter, err := gateway.NewGRPCClient(c.config)
	if err != nil {
		c.Error(err)
		return 1
	}
	params := &usecase.InteractorParams{
		Env:            env,
		OutputPort:     presenter.NewJSONCLIPresenterWithIndent(),
		GRPCPort:       grpcAdapter,
		DynamicBuilder: gateway.NewDynamicBuilder(),
	}

	var status int
	if isCommandLineMode(c.options) {
		status = c.runAsCLI(params)
	} else {
		status = c.runAsREPL(params, env)
	}

	return status
}

func (c *CLI) runAsCLI(p *usecase.InteractorParams) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // for non-zero return value

	errCh := make(chan error, 1)
	go checkUpdate(ctx, c.cache, errCh)

	var in io.Reader
	if c.options.File != "" {
		f, err := os.Open(c.options.File)
		if err != nil {
			c.Error(err)
			return 1
		}
		defer f.Close()
		in = f
	} else {
		in = os.Stdin
	}

	p.InputterPort = gateway.NewJSONFileInputter(in)
	interactor := usecase.NewInteractor(p)

	res, err := interactor.Call(&port.CallParams{c.options.Call})
	if err != nil {
		c.Error(err)
		return 1
	}

	b := new(bytes.Buffer)
	if _, err := b.ReadFrom(res); err != nil {
		c.Error(err)
		return 1
	}

	c.ui.Println(b.String())

	cancel()
	if err := <-errCh; err != nil {
		c.Error(err)
		return 1
	}
	<-ctx.Done()

	return 0
}

func (c *CLI) runAsREPL(p *usecase.InteractorParams, env *entity.Env) int {
	if c.cache.UpdateAvailable {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // for non-zero return value

		var m updater.Means
		var err error
		switch c.cache.InstalledBy {
		case meta.MeansTypeGitHubRelease:
			m, err = updater.NewMeans(github.GitHubReleaseMeans("ktr0731", "evans"))
		case meta.MeansTypeHomeBrew:
			m, err = updater.NewMeans(brew.HomeBrewMeans("ktr0731/evans", "evans"))
		}
		// if ErrUnavailable, user installed Evans by manually, ignore
		// return error response if other than errors
		if err != nil && err != updater.ErrUnavailable {
			c.Error(err)
			return 1
		}
		c.printUpdateInfo(c.cache.LatestVersion)
		var yes bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "update?",
		}, &yes, nil); err != nil {
			c.Error(err)
			return 1
		}
		if yes {
			if err := update(ctx, c.ui.Writer(), updater.New(meta.Version, m)); err != nil {
				c.Error(err)
				return 1
			}
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			defer cancel()
			<-sigCh
		}()
		<-ctx.Done()
	}

	p.InputterPort = gateway.NewPrompt(c.config, env)
	interactor := usecase.NewInteractor(p)

	var ui ui
	if c.config.REPL.ColoredOutput {
		ui = newColoredREPLUI("")
	} else {
		ui = newREPLUI("")
	}
	r := NewREPL(c.config.REPL, env, ui, interactor)
	if err := r.Start(); err != nil {
		c.Error(err)
		return 1
	}

	return 0
}

var updateInfoFormat string = `
  new update available:
    old version: %s
    new version: %s
`

func (c *CLI) printUpdateInfo(latest string) {
	c.ui.Println(fmt.Sprintf(updateInfoFormat, meta.Version, latest))
}

func checkPrecondition(config *config.Config, opt *Options) error {
	if err := isCallable(config, opt); err != nil {
		return err
	}
	return nil
}

func isCallable(config *config.Config, opt *Options) error {
	if opt.Call == "" {
		return nil
	}

	var result *multierror.Error
	if config.Default.Service == "" && opt.Service == "" {
		result = multierror.Append(result, errors.New("--service flag unselected"))
	}
	if config.Default.Package == "" && opt.Package == "" {
		result = multierror.Append(result, errors.New("--package flag unselected"))
	}
	if result != nil {
		result.ErrorFormat = func(errs []error) string {
			var txt string
			for _, e := range errs {
				txt += fmt.Sprintf("  %s\n", e)
			}
			return fmt.Sprintf("--call option needs to  options below also:\n\n%s", txt)
		}
		return result
	}
	return nil
}

func isCommandLineMode(opt *Options) bool {
	return !isatty.IsTerminal(os.Stdin.Fd()) || opt.File != ""
}

func setupEnv(conf *config.Config, opt *Options) (*entity.Env, error) {
	if len(opt.Header) > 0 {
		for _, h := range opt.Header {
			s := strings.SplitN(h, "=", 2)
			if len(s) != 2 {
				return nil, errors.New(`header must be specified "key=val" format`)
			}
			key, val := s[0], s[1]
			conf.Request.Header = append(conf.Request.Header, config.Header{Key: key, Val: val})
		}
	}
	if opt.Host != "" {
		conf.Server.Host = opt.Host
	}
	if _, err := strconv.Atoi(opt.Port); err != nil {
		return nil, errors.New(`port must be integer`)
	}
	if opt.Port != "50051" {
		conf.Server.Port = opt.Port
	}

	paths, err := collectProtoPaths(conf, opt)
	if err != nil {
		return nil, err
	}

	files, err := collectProtoFiles(conf, opt)
	if err != nil {
		return nil, err
	}

	desc, err := parser.ParseFile(files, paths)
	if err != nil {
		return nil, err
	}

	env, err := entity.NewEnv(desc, conf)
	if err != nil {
		return nil, err
	}

	// option is higher priority than config file
	pkg := opt.Package
	if pkg == "" && conf.Default.Package != "" {
		pkg = conf.Default.Package
	}

	if pkg != "" {
		if err := env.UsePackage(pkg); err != nil {
			return nil, errors.Wrapf(err, "file %s", strings.Join(opt.Proto, ", "))
		}
	}

	svc := opt.Service
	if svc == "" && conf.Default.Service != "" {
		svc = conf.Default.Service
	}

	if svc != "" {
		if err := env.UseService(svc); err != nil {
			return nil, errors.Wrapf(err, "file %s", strings.Join(opt.Proto, ", "))
		}
	}
	return env, nil
}

func collectProtoPaths(conf *config.Config, opt *Options) ([]string, error) {
	paths := make([]string, 0, len(opt.Path)+len(conf.Default.ProtoPath))
	encountered := map[string]bool{}
	parser := shellwords.NewParser()
	parser.ParseEnv = true

	parse := func(p string) (string, error) {
		res, err := parser.Parse(p)
		if err != nil {
			return "", err
		}
		if len(res) > 1 {
			return "", errors.New("failed to parse proto path")
		}
		// empty path
		if len(res) == 0 {
			return "", nil
		}
		return res[0], nil
	}

	for _, p := range append(opt.Path, conf.Default.ProtoPath...) {
		path, err := parse(p)
		if err != nil {
			return nil, err
		}

		if encountered[path] || path == "" {
			continue
		}
		encountered[path] = true
		paths = append(paths, path)
	}
	for _, proto := range opt.Proto {
		p, err := parse(proto)
		if err != nil {
			return nil, err
		}
		path := filepath.Dir(p)

		if encountered[path] || path == "" {
			continue
		}
		paths = append(paths, path)
		encountered[path] = true
	}
	return paths, nil
}

func collectProtoFiles(conf *config.Config, opt *Options) ([]string, error) {
	files := make([]string, 0, len(conf.Default.ProtoFile)+len(opt.Proto))
	for _, f := range append(conf.Default.ProtoFile, opt.Proto...) {
		if f != "" {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		return nil, ErrProtoFileRequired
	}
	return files, nil
}
