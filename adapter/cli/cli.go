package cli

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/ktr0731/evans/adapter/cui"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/di"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/evans/usecase/port"
)

// TODO: define cli mode scoped config

var DefaultReader io.Reader = os.Stdin

func Run(cfg *config.Config, ui cui.UI, file, call string) error {
	in := DefaultReader

	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	p, err := di.NewCLIInteractorParams(cfg, in)
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

	if _, err := io.Copy(ui.Writer(), res); err != nil {
		return err
	}

	return nil
}
