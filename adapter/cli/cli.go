package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ktr0731/evans/adapter/cui"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/di"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/evans/usecase/port"
	isatty "github.com/mattn/go-isatty"
)

// TODO: define cli mode scoped config

var DefaultReader io.Reader = os.Stdin

type LaunchError struct {
	err error
}

func (e *LaunchError) Error() string {
	return fmt.Sprintf("failed to launch CLI mode: %s", e.err)
}

// Run is a main entrypoint for CLI mode.
// cli package will executes Run if Evans is launched as CLI mode.
//
// Run returns below errors (Clients should unwrap returned error with errors.Cause):
// - os.PathError
//   - Provided `file` is missing.
//   - It is returned only when `file` is not empty.
// - ErrLaunchFailed
//   - Precondition error to launch CLI mode.
// - TODO: Describe more error specification.
//
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
		return &LaunchError{err}
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

// IsCLIMode returns whether Evans was launched as CLI mode or not.
func IsCLIMode(file string) bool {
	return !isatty.IsTerminal(os.Stdin.Fd()) || file != ""
}
