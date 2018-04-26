package controller

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/AlecAivazis/survey"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/ktr0731/evans/adapter/gateway"
	"github.com/ktr0731/evans/adapter/parser"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/evans/usecase/port"
	updater "github.com/ktr0731/go-updater"
	isatty "github.com/mattn/go-isatty"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/pkg/errors"

	"io"
	"os"
)

var (
	ErrProtoFileRequired = errors.New("least one proto file required")
)

type optStrSlice []string

func (o *optStrSlice) String() string {
	return fmt.Sprintf("%v", *o)
}

func (o *optStrSlice) Set(v string) error {
	*o = append(*o, v)
	return nil
}

var usageFormat = `
Usage: %s [options ...] [PROTO [PROTO ...]]

Positional arguments:
	PROTO			.proto files

Options:
	--interactive, -i	%s
	--edit, -e		%s
	--host HOST		%s
	--port PORT, -p PORT	%s
	--package PACKAGE	%s
	--service SERVICE	%s
	--call CALL		%s
	--file FILE, -f FILE	%s
	--path PATH		%s
	--header HEADER		%s
	--help, -h		%s
	--version, -v		%s
`

func (c *CLI) parseFlags(args []string) {
	const (
		interactive = "use interactive mode"
		edit        = "edit config file using by $EDITOR"
		host        = "gRPC server host"
		port        = "gRPC server port"
		pkg         = "default package"
		service     = "default service"
		call        = "call specified RPC by CLI mode"
		file        = "the script file which will be executed by (used only CLI mode)"
		path        = "proto file paths"
		header      = "default headers which set to each requests (example: foo=bar)"

		version = "display version and exit"
		help    = "display this help and exit"
	)

	f := flag.NewFlagSet("main", flag.ExitOnError)
	f.Usage = func() {
		c.Version()
		fmt.Fprintf(
			c.ui.Writer(),
			usageFormat,
			c.name,
			interactive,
			edit,
			host,
			port,
			pkg,
			service,
			call,
			file,
			path,
			header,
			version,
			help,
		)
		os.Exit(0)
	}

	f.BoolVar(&c.options.Interactive, "interactive", false, interactive)
	f.BoolVar(&c.options.Interactive, "i", false, interactive)
	f.BoolVar(&c.options.EditConfig, "edit", false, edit)
	f.BoolVar(&c.options.EditConfig, "e", false, edit)
	f.StringVar(&c.options.Host, "host", "", host)
	f.StringVar(&c.options.Port, "port", "50051", port)
	f.StringVar(&c.options.Port, "p", "50051", port)
	f.StringVar(&c.options.Package, "package", "", pkg)
	f.StringVar(&c.options.Service, "service", "", service)
	f.StringVar(&c.options.Call, "call", "", call)
	f.StringVar(&c.options.File, "file", "", file)
	f.StringVar(&c.options.File, "f", "", file)
	f.Var(&c.options.Path, "path", path)
	f.Var(&c.options.Header, "header", header)
	f.BoolVar(&c.options.version, "version", false, version)
	f.BoolVar(&c.options.version, "v", false, version)

	// ignore error because flag set mode is ExitOnError
	_ = f.Parse(args)

	c.flagSet = f
}

type Options struct {
	Interactive bool
	EditConfig  bool
	Host        string
	Port        string
	Package     string
	Service     string
	Call        string
	File        string
	Path        optStrSlice
	Header      optStrSlice

	version bool
	help    bool
}

type CLI struct {
	name    string
	version string

	ui     UI
	config *config.Config

	flagSet *flag.FlagSet
	options *Options

	cache *cache.Cache
}

// NewCLI instantiate CLI interface.
// if Evans is used as REPL mode, its UI is created by newREPLUI() in runAsREPL.
// if CLI mode, its ui is same as passed ui.
func NewCLI(name, version string, ui UI) *CLI {
	return &CLI{
		name:    name,
		version: version,
		options: &Options{},
		ui:      ui,
		config:  config.Get(),
		cache:   cache.Get(),
	}
}

func (c *CLI) Error(err error) {
	c.ui.ErrPrintln(err.Error())
}

func (c *CLI) Usage() {
	c.flagSet.Usage()
}

func (c *CLI) Version() {
	c.ui.Println(fmt.Sprintf("%s %s", c.name, c.version))
}

func (c *CLI) Run(args []string) int {
	c.parseFlags(args)
	proto := c.flagSet.Args()

	switch {
	case c.options.version:
		c.Version()
		return 0
	case c.options.help:
		c.Usage()
		return 0
	}

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

	env, err := setupEnv(c.config, c.options, proto)
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
	go checkUpdate(ctx, c.config, c.cache, errCh)

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// if AutoUpdate enabled, do update asynchronously
	puCh := make(chan error, 1)
	if c.config.Meta.AutoUpdate {
		go func() {
			puCh <- c.processUpdate(ctx)
		}()
	} else {
		err := c.processUpdate(ctx)
		if err != nil {
			c.Error(err)
			return 1
		}
	}

	cuCh := make(chan error, 1)
	go checkUpdate(ctx, c.config, c.cache, cuCh)

	p.InputterPort = gateway.NewPrompt(c.config, env)
	interactor := usecase.NewInteractor(p)

	var ui UI
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

	cancel()
	<-ctx.Done()
	if c.config.Meta.AutoUpdate {
		if err := <-puCh; err != nil {
			c.Error(err)
			return 1
		}
	}
	if err := <-cuCh; err != nil {
		c.Error(err)
		return 1
	}

	return 0
}

// processUpdate checks new changes and updates Evans in accordance with user's selection.
// if config.Meta.AutoUpdate enabled, processUpdate is called asynchronously.
// other than, processUpdate is called synchronously.
func (c *CLI) processUpdate(ctx context.Context) error {
	if !c.cache.UpdateAvailable {
		return nil
	}

	m, err := newMeans(c.cache)
	// if ErrUnavailable, user installed Evans by manually, ignore
	if err == updater.ErrUnavailable {
		// show update info at the end
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "failed to get means from cache (%s)", c.cache)
	}

	var w io.Writer
	if c.config.Meta.AutoUpdate {
		w = ioutil.Discard

		// if canceled, ignore and return
		err := update(ctx, w, newUpdater(c.config, meta.Version, m))
		if errors.Cause(err) == context.Canceled {
			return nil
		}
		return err
	}

	printUpdateInfo(c.ui.Writer(), c.cache.LatestVersion)

	var yes bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "update?",
	}, &yes, nil); err != nil {
		return errors.Wrap(err, "failed to get survey answer")
	}
	if !yes {
		return nil
	}

	w = c.ui.Writer()

	// if canceled, ignore and return
	err = update(ctx, w, newUpdater(c.config, meta.Version, m))
	if errors.Cause(err) == context.Canceled {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "failed to update binary")
	}

	// restart Evans
	if err := syscall.Exec(os.Args[0], os.Args, os.Environ()); err != nil {
		return errors.Wrapf(err, "failed to exec the command: args=%s", os.Args)
	}

	return nil
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

func setupEnv(conf *config.Config, opt *Options, proto []string) (*entity.Env, error) {
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

	paths, err := collectProtoPaths(conf, opt, proto)
	if err != nil {
		return nil, err
	}

	files, err := collectProtoFiles(conf, opt, proto)
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
			return nil, errors.Wrapf(err, "file %s", strings.Join(proto, ", "))
		}
	}

	svc := opt.Service
	if svc == "" && conf.Default.Service != "" {
		svc = conf.Default.Service
	}

	if svc != "" {
		if err := env.UseService(svc); err != nil {
			return nil, errors.Wrapf(err, "file %s", strings.Join(proto, ", "))
		}
	}
	return env, nil
}

func collectProtoPaths(conf *config.Config, opt *Options, proto []string) ([]string, error) {
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
	for _, proto := range proto {
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

func collectProtoFiles(conf *config.Config, opt *Options, proto []string) ([]string, error) {
	files := make([]string, 0, len(conf.Default.ProtoFile)+len(proto))
	for _, f := range append(conf.Default.ProtoFile, proto...) {
		if f != "" {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		return nil, ErrProtoFileRequired
	}
	return files, nil
}
