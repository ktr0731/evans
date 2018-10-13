package cmd

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
	"github.com/ktr0731/evans/adapter/cli"
	"github.com/ktr0731/evans/adapter/controller"
	"github.com/ktr0731/evans/adapter/cui"
	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/di"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/evans/usecase"
	semver "github.com/ktr0731/go-semver"
	updater "github.com/ktr0731/go-updater"
	isatty "github.com/mattn/go-isatty"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"io"
	"os"
)

var (
	ErrProtoFileRequired = errors.New("least one proto file required")
)

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

	ui   cui.CUI
	wcfg *wrappedConfig

	flagSet *flag.FlagSet

	cache *cache.Cache

	initOnce sync.Once
}

// New instantiate a Command.
// Evans accepts some options and args and selects either REPL mode or CLI mode.
func New(name, version string, ui cui.CUI) *Command {
	return &Command{
		name:    name,
		version: version,
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

func (c *Command) Error(err error) {
	c.ui.ErrPrintln(err.Error())
}

func (c *Command) Usage() {
	c.flagSet.Usage()
}

func (c *Command) Version() {
	c.ui.Println(fmt.Sprintf("%s %s", c.name, c.version))
}

func (c *Command) Run(args []string) int {
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

	if err := c.init(opts, proto); err != nil {
		c.Error(err)
		return 1
	}

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

func (c *Command) runAsCLI() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // for non-zero return value

	var eg errgroup.Group

	checkUpdateErrCh := make(chan error, 1)
	eg.Go(func() error {
		return checkUpdate(ctx, c.wcfg.cfg, c.cache)
	})

	errCh := make(chan error)
	eg.Go(func() error {
		defer cancel()
		return cli.New(c.ui, c.wcfg.cfg).Run(ctx, c.wcfg.file, c.wcfg.call)
	})

	go func() {
		errCh <- eg.Wait()
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
var DefaultREPLUI = controller.NewREPLUI("")
var DefaultCLIReader io.Reader = os.Stdin

func (c *Command) runAsREPL() int {
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
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer closeCancel()
		defer p.Cleanup(closeCtx)

		interactor := usecase.NewInteractor(p)

		var ui controller.UI
		if c.wcfg.cfg.REPL.ColoredOutput {
			ui = controller.NewColoredREPLUI(DefaultREPLUI)
		} else {
			ui = DefaultREPLUI
		}

		env, err := di.Env(c.wcfg.cfg)
		if err != nil {
			errCh <- err
			return
		}

		r := controller.NewREPL(c.wcfg.cfg.REPL, env, ui, interactor)
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
	if w.cli && w.call == "" {
		return errors.New("--cli flag needs to --call flag")
	} else if w.call == "" {
		// REPL mode
		return nil
	}

	var result *multierror.Error
	if w.cfg.Default.Service == "" {
		result = multierror.Append(result, errors.New("--service flag unselected"))
	}
	if !w.cfg.Server.Reflection && w.cfg.Default.Package == "" {
		result = multierror.Append(result, errors.New("--package flag unselected"))
	}
	if result != nil {
		result.ErrorFormat = func(errs []error) string {
			var txt string
			for _, e := range errs {
				txt += fmt.Sprintf("  %s\n", e)
			}
			return fmt.Sprintf("--call option needs to options below also:\n\n%s", txt)
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
