package cli

import (
	"context"
	"time"

	"github.com/ktr0731/evans/adapter/cui"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/di"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/evans/usecase/port"

	"io"
	"os"
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

type CLI struct {
	ui  cui.CUI
	cfg *config.Config
}

// New instantiate CLI interface.
// if Evans is used as REPL mode, its UI is created by newREPLUI() in runAsREPL.
// if CLI mode, its ui is same as passed ui.
func New(ui cui.CUI, cfg *config.Config) *CLI {
	return &CLI{
		ui:  ui,
		cfg: cfg,
	}
}

var DefaultCLIReader io.Reader = os.Stdin

// TODO: define CLI mode specific config type instead of args.

func (c *CLI) Run(ctx context.Context, file, call string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // for non-zero return value

	in := DefaultCLIReader
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	p, err := di.NewCLIInteractorParams(c.cfg, in)
	if err != nil {
		return err
	}
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer closeCancel()
	defer p.Cleanup(closeCtx)

	interactor := usecase.NewInteractor(p)

	res, err := interactor.Call(&port.CallParams{RPCName: call})
	if err != nil {
		return err
	}

	if _, err := io.Copy(c.ui.Writer(), res); err != nil {
		return err
	}

	return nil
}
