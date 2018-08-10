package controller

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/di"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/evans/usecase/port"
	semver "github.com/ktr0731/go-semver"
	updater "github.com/ktr0731/go-updater"
	isatty "github.com/mattn/go-isatty"
	"github.com/mitchellh/copystructure"
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
	--edit, -e		%s
	--repl			%s
	--cli			%s
	--silent, -s		%s
	--host HOST		%s
	--port PORT, -p PORT	%s
	--package PACKAGE	%s
	--service SERVICE	%s
	--call CALL		%s
	--file FILE, -f FILE	%s
	--path PATH		%s
	--header HEADER		%s
	--web			%s
	--reflection, -r	%s

	--help, -h		%s
	--version, -v		%s
`

func (c *CLI) parseFlags(args []string) *options {
	const (
		edit       = "edit config file using by $EDITOR"
		repl       = "start as REPL mode"
		cli        = "start as CLI mode"
		silent     = "hide splash"
		host       = "gRPC server host"
		port       = "gRPC server port"
		pkg        = "default package"
		service    = "default service"
		call       = "call specified RPC by CLI mode"
		file       = "the script file which will be executed by (used only CLI mode)"
		path       = "proto file paths"
		header     = "default headers which set to each requests (example: foo=bar)"
		web        = "use gRPC Web protocol"
		reflection = "use gRPC reflection"

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
			edit,
			repl,
			cli,
			silent,
			host,
			port,
			pkg,
			service,
			call,
			file,
			path,
			header,
			web,
			reflection,
			version,
			help,
		)
	}

	var opts options

	f.BoolVar(&opts.editConfig, "edit", false, edit)
	f.BoolVar(&opts.editConfig, "e", false, edit)
	f.BoolVar(&opts.repl, "repl", false, repl)
	f.BoolVar(&opts.cli, "cli", false, cli)
	f.BoolVar(&opts.silent, "silent", false, silent)
	f.BoolVar(&opts.silent, "s", false, silent)
	f.StringVar(&opts.host, "host", "", host)
	f.StringVar(&opts.port, "port", "50051", port)
	f.StringVar(&opts.port, "p", "50051", port)
	f.StringVar(&opts.pkg, "package", "", pkg)
	f.StringVar(&opts.service, "service", "", service)
	f.StringVar(&opts.call, "call", "", call)
	f.StringVar(&opts.file, "file", "", file)
	f.StringVar(&opts.file, "f", "", file)
	f.Var(&opts.path, "path", path)
	f.Var(&opts.header, "header", header)
	f.BoolVar(&opts.web, "web", false, web)
	f.BoolVar(&opts.reflection, "reflection", false, reflection)
	f.BoolVar(&opts.reflection, "r", false, reflection)
	f.BoolVar(&opts.version, "version", false, version)
	f.BoolVar(&opts.version, "v", false, version)

	// ignore error because flag set mode is ExitOnError
	_ = f.Parse(args)

	c.flagSet = f

	return &opts
}

type options struct {
	// mode options
	editConfig bool

	// config options
	repl       bool
	cli        bool
	silent     bool
	host       string
	port       string
	pkg        string
	service    string
	call       string
	file       string
	path       optStrSlice
	header     optStrSlice
	web        bool
	reflection bool

	// meta options
	version bool
}

// wrappedConfig is created at intialization and
// it has *config.Config and other fields.
// *config.Config is a merged config by mergeConfig.
// other fields will be copied by c.init.
// these fields are belong to options, but not config.Config
// like call field.
type wrappedConfig struct {
	cfg *config.Config

	// used only CLI mode
	call string
	// used as a input for CLI mode
	// if input is stdin, file is empty
	file string

	// explicit using REPL mode
	repl bool

	// explicit using CLI mode
	cli bool
}

type CLI struct {
	name    string
	version string

	ui   UI
	wcfg *wrappedConfig

	flagSet *flag.FlagSet

	cache *cache.Cache

	initOnce sync.Once
}

// NewCLI instantiate CLI interface.
// if Evans is used as REPL mode, its UI is created by newREPLUI() in runAsREPL.
// if CLI mode, its ui is same as passed ui.
func NewCLI(name, version string, ui UI) *CLI {
	return &CLI{
		name:    name,
		version: version,
		ui:      ui,
		cache:   cache.Get(),
	}
}

func (c *CLI) init(opts *options, proto []string) error {
	var err error
	c.initOnce.Do(func() {
		var cfg *config.Config
		cfg, err = mergeConfig(config.Get(), opts, proto)
		if err != nil {
			return
		}

		c.wcfg = &wrappedConfig{
			cfg:  cfg,
			call: opts.call,
			file: opts.file,
			repl: opts.repl,
			cli:  opts.cli,
		}

		err = checkPrecondition(c.wcfg)
		if err != nil {
			return
		}
	})
	return err
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
	opts := c.parseFlags(args)
	proto := c.flagSet.Args()

	switch {
	case opts.version:
		c.Version()
		return 0
	case opts.editConfig:
		if err := config.Edit(); err != nil {
			c.Error(err)
			return 1
		}
		return 0
	}

	c.init(opts, proto)

	if len(c.wcfg.cfg.Default.ProtoFile) == 0 && !c.wcfg.cfg.Server.Reflection {
		c.Usage()
		c.Error(ErrProtoFileRequired)
		return 1
	}

	var status int
	if isCommandLineMode(c.wcfg) {
		status = c.runAsCLI()
	} else {
		status = c.runAsREPL()
	}

	return status
}

var DefaultCLIReader io.Reader = os.Stdin

func (c *CLI) runAsCLI() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // for non-zero return value

	checkUpdateErrCh := make(chan error, 1)
	go func() {
		checkUpdateErrCh <- checkUpdate(ctx, c.wcfg.cfg, c.cache)
	}()

	errCh := make(chan error)
	go func() {
		defer cancel()

		in := DefaultCLIReader
		if c.wcfg.file != "" {
			f, err := os.Open(c.wcfg.file)
			if err != nil {
				errCh <- err
				return
			}
			defer f.Close()
			in = f
		}

		p, err := di.NewCLIInteractorParams(c.wcfg.cfg, in)
		if err != nil {
			errCh <- err
			return
		}
		interactor := usecase.NewInteractor(p)
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer closeCancel()
		defer interactor.Close(closeCtx)

		res, err := interactor.Call(&port.CallParams{RPCName: c.wcfg.call})
		if err != nil {
			errCh <- err
			return
		}

		if _, err := io.Copy(c.ui.Writer(), res); err != nil {
			errCh <- err
			return
		}
	}()

	select {
	case <-ctx.Done():
		return 0
	case err := <-errCh:
		// first, cancel
		cancel()

		// receive the REPL result
		if err != nil {
			c.Error(err)
		}

		cuErr := <-checkUpdateErrCh
		if cuErr != nil {
			c.Error(cuErr)
		}

		if err != nil || cuErr != nil {
			return 1
		}
	}
	return 0
}

// DefaultREPLUI is used for e2e testing
var DefaultREPLUI = newREPLUI("")

func (c *CLI) runAsREPL() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	checkUpdateErrCh := make(chan error, 1)
	go func() {
		checkUpdateErrCh <- checkUpdate(ctx, c.wcfg.cfg, c.cache)
	}()

	processUpdateErrCh := make(chan error, 1)
	errCh := make(chan error)
	go func() {
		defer cancel()

		// if AutoUpdate enabled, do update asynchronously
		if c.wcfg.cfg.Meta.AutoUpdate {
			go func() {
				processUpdateErrCh <- c.processUpdate(ctx)
			}()
		} else {
			err := c.processUpdate(ctx)
			if err != nil {
				errCh <- err
				return
			}
		}

		p, err := di.NewREPLInteractorParams(c.wcfg.cfg, DefaultCLIReader)
		if err != nil {
			errCh <- err
			return
		}
		interactor := usecase.NewInteractor(p)
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer closeCancel()
		defer interactor.Close(closeCtx)

		var ui UI
		if c.wcfg.cfg.REPL.ColoredOutput {
			ui = newColoredREPLUI(DefaultREPLUI)
		} else {
			ui = DefaultREPLUI
		}

		env, err := di.Env(c.wcfg.cfg)
		if err != nil {
			errCh <- err
			return
		}

		r := NewREPL(c.wcfg.cfg.REPL, env, ui, interactor)
		if err := r.Start(); err != nil {
			errCh <- err
			return
		}
	}()

	select {
	case <-ctx.Done():
		return 0
	case err := <-errCh:
		if err != nil {
			c.Error(err)
			return 1
		}

		cancel()

		select {
		case err = <-processUpdateErrCh:
		case err = <-checkUpdateErrCh:
		}

		if err != nil {
			c.Error(err)
			return 1
		}
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

	v := semver.MustParse(c.cache.LatestVersion)
	if v.LessThan(meta.Version) || v.Equal(meta.Version) {
		return cache.Clear()
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
	if c.wcfg.cfg.Meta.AutoUpdate {
		w = ioutil.Discard

		// if canceled, ignore and return
		err := update(ctx, w, newUpdater(c.wcfg.cfg, meta.Version, m))
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
	err = update(ctx, w, newUpdater(c.wcfg.cfg, meta.Version, m))
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

func mergeConfig(cfg *config.Config, opt *options, proto []string) (*config.Config, error) {
	headers, err := toHeader(opt.header)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge config and option")
	}

	mergeString := func(s1, s2 string) string {
		if s2 != "" {
			return s2
		}
		return s1
	}

	mergeSlice := func(s1, s2 []string) []string {
		slice := make([]string, 0, len(s1)+len(s2))
		encountered := map[string]bool{}
		for _, s := range append(s1, s2...) {
			if !encountered[s] {
				slice = append(slice, s)
				encountered[s] = true
			}
		}
		return slice
	}

	mergeHeader := func(s1, s2 []config.Header) []config.Header {
		slice := make([]config.Header, 0, len(s1)+len(s2))
		encountered := map[string]bool{}
		for _, s := range append(s1, s2...) {
			if !encountered[s.Key] {
				slice = append(slice, s)
				encountered[s.Key] = true
			}
		}
		return slice
	}

	mc := copystructure.Must(copystructure.Copy(cfg)).(*config.Config)

	mc.Default.Package = mergeString(cfg.Default.Package, opt.pkg)
	mc.Default.Service = mergeString(cfg.Default.Service, opt.service)
	mc.Default.ProtoPath = mergeSlice(cfg.Default.ProtoPath, opt.path)
	mc.Default.ProtoFile = mergeSlice(cfg.Default.ProtoFile, proto)

	mc.Server.Host = mergeString(cfg.Server.Host, opt.host)
	mc.Server.Port = mergeString(cfg.Server.Port, opt.port)

	mc.Request.Header = mergeHeader(cfg.Request.Header, headers)

	if opt.silent {
		mc.REPL.ShowSplashText = false
	}

	if opt.web {
		mc.Request.Web = true
	}

	if opt.reflection {
		mc.Server.Reflection = true
	}

	config.SetupConfig(mc)
	return mc, nil
}

func checkPrecondition(w *wrappedConfig) error {
	if _, err := strconv.Atoi(w.cfg.Server.Port); err != nil {
		return errors.New(`port must be integer`)
	}

	if err := isCallable(w); err != nil {
		return errors.Wrap(err, "not callable")
	}

	if w.cli && w.repl {
		return errors.New("cannot use both of --cli and --repl options")
	}

	if w.cfg.Server.Reflection && w.cfg.Request.Web {
		return errors.New("gRPC Web server reflection is not supported yet")
	}

	return nil
}

func isCallable(w *wrappedConfig) error {
	if w.call == "" {
		return nil
	}

	var result *multierror.Error
	if w.cfg.Default.Service == "" {
		result = multierror.Append(result, errors.New("--service flag unselected"))
	}
	if w.cfg.Default.Package == "" {
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

func isCommandLineMode(w *wrappedConfig) bool {
	return !w.repl && (!isatty.IsTerminal(os.Stdin.Fd()) || w.file != "")
}

func toHeader(sh optStrSlice) ([]config.Header, error) {
	if len(sh) == 0 {
		return nil, nil
	}
	headers := make([]config.Header, 0, len(sh))
	for _, h := range sh {
		s := strings.SplitN(h, "=", 2)
		if len(s) != 2 {
			return nil, errors.New(`header must be specified "key=val" format`)
		}
		headers = append(headers, config.Header{
			Key: s[0],
			Val: s[1],
		})
	}
	return headers, nil
}
