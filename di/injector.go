// +build wireinject

package di

import (
	"github.com/google/go-cloud/wire"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/usecase"
)

func initializeCLIParams(cfg *config.Config, in JSONReader) (*usecase.InteractorParams, error) {
	wire.Build(CLIParams)
	return &usecase.InteractorParams{}, nil
}
