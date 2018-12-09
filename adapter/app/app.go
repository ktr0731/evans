package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"

	"github.com/AlecAivazis/survey"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/ktr0731/evans/adapter/cli"
	"github.com/ktr0731/evans/adapter/cui"
	"github.com/ktr0731/evans/adapter/repl"
	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/meta"
	semver "github.com/ktr0731/go-semver"
	updater "github.com/ktr0731/go-updater"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

var (
	ErrProtoFileRequired = errors.New("least one proto file required")
)

var usageFormat = `
Usage: %s [--help] [--version] [options ...] [PROTO [PROTO ...]]

Positional arguments:
	PROTO			.proto files

Options:
%s
`

func (c *Command) parseFlags(args []string) *options {
	f := pflag.NewFlagSet("main", pflag.ExitOnError)
	f.SortFlags = false
	f.SetOutput(c.ui.Writer())

	var opts options

	f.BoolVarP(&opts.editConfig, "edit", "e", false, "edit config file using by $EDITOR")
	f.BoolVar(&opts.repl, "repl", false, "launch Evans as REPL mode")
	f.BoolVar(&opts.cli, "cli", false, "start as CLI mode")
	f.BoolVarP(&opts.silent, "silent", "s", false, "hide splash")
	f.StringVar(&opts.host, "host", "", "gRPC server host")
	f.StringVarP(&opts.port, "port", "p", "50051", "gRPC server port")
	f.StringVar(&opts.pkg, "package", "", "default package")
	f.StringVar(&opts.service, "service", "", "default service")
	f.StringVar(&opts.call, "call", "", "call specified RPC by CLI mode")
	f.StringVarP(&opts.file, "file", "f", "", "a script file that will be executed by (used only CLI mode)")
	f.StringSliceVar(&opts.path, "path", nil, "proto file paths")
	f.StringSliceVar(&opts.header, "header", nil, "default headers that set to each requests (example: foo=bar)")
	f.BoolVar(&opts.web, "web", false, "use gRPC Web protocol")
	f.BoolVarP(&opts.reflection, "reflection", "r", false, "use gRPC reflection")
	f.BoolVarP(&opts.version, "version", "v", false, "display version and exit")
	f.BoolVarP(&opts.help, "help", "h", false, "display help text and exit")

	f.Usage = func() {
		c.printVersion()
		var buf bytes.Buffer
		w := tabwriter.NewWriter(&buf, 0, 8, 8, ' ', tabwriter.TabIndent)
		f.VisitAll(func(f *pflag.Flag) {
			cmd := "--" + f.Name
			if f.Shorthand != "" {
				cmd += ", -" + f.Shorthand
			}
			name, _ := pflag.UnquoteUsage(f)
			if name != "" {
				cmd += " " + name
			}
			usage := f.Usage
			if f.DefValue != "" {
				usage += fmt.Sprintf(` (default "%s")`, f.DefValue)
			}
			fmt.Fprintf(w, "        %s\t%s\t\n", cmd, usage)
		})
		w.Flush()
		fmt.Fprintf(c.ui.Writer(), usageFormat, meta.AppName, buf.String())
	}

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
	path       []string
	header     []string
	web        bool
	reflection bool

	// meta options
	version bool
	help    bool
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

type Command struct {
	name    string
	version string

	ui   cui.UI
	wcfg *wrappedConfig

	flagSet *pflag.FlagSet

	cache *cache.Cache

	initOnce sync.Once
}

// New instantiate CLI interface.
// ui is used for both of CLI mode and REPL mode.
// If ui is nil, cui.NewBasicUI will be used.
func New(ui cui.UI) *Command {
	if ui == nil {
		ui = cui.NewBasic()
	}
	return &Command{
		name:    meta.AppName,
		version: meta.Version.String(),
		ui:      ui,
		cache:   cache.Get(),
	}
}

func (c *Command) init(opts *options, proto []string) error {
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

func (c *Command) printUsage() {
	c.flagSet.Usage()
}

func (c *Command) printVersion() {
	c.ui.Println(fmt.Sprintf("%s %s", c.name, c.version))
}

// Run starts Evans.
// If returned int value is 0, Evans has finished normally.
// Conversely value is 1, Evans has finished with some errors.
func (c *Command) Run(args []string) int {
	err := c.run(args)
	if err != nil {
		c.ui.ErrPrintln(err.Error())
		return 1
	}
	return 0
}

func (c *Command) run(args []string) error {
	opts := c.parseFlags(args)
	proto := c.flagSet.Args()

	switch {
	case opts.version:
		c.printVersion()
		return nil
	case opts.help:
		c.flagSet.Usage()
		return nil
	case opts.editConfig:
		if err := config.Edit(); err != nil {
			return err
		}
		return nil
	}

	c.init(opts, proto)

	if len(c.wcfg.cfg.Default.ProtoFile) == 0 && !c.wcfg.cfg.Server.Reflection {
		c.printUsage()
		return ErrProtoFileRequired
	}

	var err error
	// TODO: use c.wcfg.cli instead of c.wcfg.repl
	if !c.wcfg.repl && cli.IsCLIMode(c.wcfg.file) {
		err = c.runAsCLI()
	} else {
		err = c.runAsREPL()
	}

	return err
}

func (c *Command) runAsCLI() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // for non-zero return value

	checkUpdateErrCh := make(chan error, 1)
	go func() {
		checkUpdateErrCh <- checkUpdate(ctx, c.wcfg.cfg, c.cache)
	}()

	err := cli.Run(c.wcfg.cfg, c.ui, c.wcfg.file, c.wcfg.call)
	if err != nil {
		return err
	}

	cancel()
	<-ctx.Done()

	return nil
}

func (c *Command) runAsREPL() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	checkUpdateErrCh := make(chan error, 1)
	go func() {
		checkUpdateErrCh <- checkUpdate(ctx, c.wcfg.cfg, c.cache)
	}()

	processUpdateErrCh := make(chan error, 1)
	// if AutoUpdate enabled, do update asynchronously
	if c.wcfg.cfg.Meta.AutoUpdate {
		go func() {
			processUpdateErrCh <- c.processUpdate(ctx)
		}()
	} else {
		err := c.processUpdate(ctx)
		if err != nil {
			return err
		}
	}

	err := repl.Run(c.wcfg.cfg, c.ui)
	if err != nil {
		return err
	}

	cancel()

	select {
	case <-ctx.Done():
		return nil
	case err := <-checkUpdateErrCh:
		if err != nil {
			return err
		}
	case err := <-processUpdateErrCh:
		if err != nil {
			return err
		}
	}

	return nil
}

// processUpdate checks new changes and updates Evans in accordance with user's selection.
// if config.Meta.AutoUpdate enabled, processUpdate is called asynchronously.
// other than, processUpdate is called synchronously.
func (c *Command) processUpdate(ctx context.Context) error {
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
		return errors.Wrapf(err, "failed to get means from cache (%v)", c.cache)
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

func toHeader(sh []string) ([]config.Header, error) {
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
