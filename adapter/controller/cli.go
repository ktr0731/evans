package controller

import (
	"fmt"
	"path/filepath"
	"strings"

	arg "github.com/alexflint/go-arg"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/ktr0731/evans/adapter/gateway"
	"github.com/ktr0731/evans/adapter/logger"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/parser"
	"github.com/ktr0731/evans/repl"
	"github.com/ktr0731/evans/usecase"
	isatty "github.com/mattn/go-isatty"
	"github.com/pkg/errors"

	"io"
	"os"
)

type UI struct {
	Reader            io.Reader
	Writer, ErrWriter io.Writer
}

func newUI() *UI {
	return &UI{
		Reader:    os.Stdin,
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
	}
}

type Options struct {
	Proto []string `arg:"positional,help:.proto files"`

	Interactive bool     `arg:"-i,help:use interactive mode"`
	EditConfig  bool     `arg:"-e,help:edit config file by $EDITOR"`
	Host        string   `arg:"-h,help:gRPC host"`
	Port        int      `arg:"-p,help:gRPC port"`
	Package     string   `arg:"help:default package"`
	Service     string   `arg:"help:default service. evans parse package from this if --package is nothing."`
	Call        string   `arg:"-c,help:call specified RPC"`
	File        string   `arg:"-f,help:the script file which will be executed (used only command-line mode)"`
	Path        []string `arg:"separate,help:proto file path"`
	Header      []string `arg:"separate,help:headers set to each requests"`
}

func (o *Options) Version() string {
	return "evans 0.1.0"
}

type CLI struct {
	ui     *UI
	config *config.Config

	parser  *arg.Parser
	options *Options
}

func NewCLI(title, version string) *CLI {
	return &CLI{
		ui: newUI(),
		options: &Options{
			Port: 50051,
		},
		config: config.Get(),
	}
}

func (c *CLI) Error(err error) {
	fmt.Fprintln(c.ui.ErrWriter, err)
}

func (c *CLI) Help() {
	c.parser.WriteHelp(c.ui.Writer)
}

func (c *CLI) Usage() {
	c.parser.WriteUsage(c.ui.Writer)
}

func (c *CLI) Run(args []string) int {
	params := &usecase.InteractorParams{
		OutputPort: presenter.NewJSONCLIPresenter(),
		Logger:     logger.NewStdLogger(config.Get()),
	}

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
	if err != nil {
		c.Error(err)
		return 1
	}

	if isCommandLineMode(c.options) {
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

		params.InputterPort = gateway.NewJSONFileInputter(in)
		usecase.NewInteractor(params)

		panic("cli mode is not implemented yet")

		if err := env.CallWithScript(in, c.options.Call); err != nil {
			c.Error(err)
			return 1
		}
	} else {
		params.InputterPort = gateway.NewPromptInputter(env)
		params.Logger = logger.NewPromptLogger()

		interactor := usecase.NewInteractor(params)
		r := repl.NewREPL(c.config.REPL, env, repl.NewBasicUI(), interactor)
		defer r.Close()

		if err := r.Start(); err != nil {
			c.Error(err)
			return 1
		}

	}
	return 0
}

func checkPrecondition(config *config.Config, opt *Options) error {
	if len(opt.Proto) == 0 {
		return errors.New("invalid argument")
	}
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
			conf.Env.Request.Header = append(conf.Env.Request.Header, config.Header{key, val})
		}
	}
	if opt.Host != "" {
		conf.Server.Host = opt.Host
	}
	if opt.Port != 50051 {
		conf.Server.Port = string(opt.Port)
	}

	// find all proto paths
	paths := make([]string, 0, len(opt.Path))
	encountered := map[string]bool{}
	for _, p := range opt.Path {
		encountered[p] = true
		paths = append(paths, p)
	}
	for _, proto := range opt.Proto {
		p := filepath.Dir(proto)
		if !encountered[p] {
			paths = append(paths, p)
			encountered[p] = true
		}
	}

	desc, err := parser.ParseFile(opt.Proto, paths)
	if err != nil {
		return nil, err
	}

	env, err := entity.New(desc, conf.Env)
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
