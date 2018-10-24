package di

import (
	"io"

	"github.com/google/go-cloud/wire"
	"github.com/ktr0731/evans/adapter/gateway"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/evans/usecase/port"
)

var CLIParams = wire.NewSet(
	provideEnv,
	provideJSONPresenter,
	provideJSONInputter,
	provideGRPCClient,
	provideDynamicBuilder,
	wire.Bind(new(port.OutputPort), new(presenter.JSONPresenter)),
	wire.Bind(new(port.Inputter), new(gateway.JSONFileInputter)),
	wire.Bind(new(port.DynamicBuilder), new(gateway.DynamicBuilder)),
	usecase.InteractorParams{},
)

func NewCLIInteractorParams(cfg *config.Config, in io.Reader) (*usecase.InteractorParams, error) {
	return initializeCLIParams(cfg, JSONReader(in))
}

// func NewCLIInteractorParams(cfg *config.Config, in io.Reader) (*usecase.InteractorParams, error) {
// 	if err := initDependencies(cfg, in); err != nil {
// 		return nil, err
// 	}
//
// 	return &usecase.InteractorParams{
// 		Env:            env,
// 		OutputPort:     jsonCLIPresenter,
// 		InputterPort:   jsonFileInputter,
// 		GRPCClient:     gRPCClient,
// 		DynamicBuilder: dynamicBuilder,
// 	}, nil
// }

func NewREPLInteractorParams(cfg *config.Config, in io.Reader) (param *usecase.InteractorParams, err error) {
	if err := initDependencies(cfg, in); err != nil {
		return nil, err
	}
	return &usecase.InteractorParams{
		Env:            env,
		OutputPort:     jsonCLIPresenter,
		InputterPort:   promptInputter,
		GRPCClient:     gRPCClient,
		DynamicBuilder: dynamicBuilder,
	}, nil
}
