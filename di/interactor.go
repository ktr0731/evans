package di

import (
	"github.com/ktr0731/evans/adapter/gateway"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/evans/usecase/port"
	multierror "github.com/ktr0731/go-multierror"
)

func NewCLIInteractorParams(cfg *config.Config, env *entity.Env, inputter port.Inputter) (*usecase.InteractorParams, error) {
	return newInteractorParams(cfg, env, inputter)
}

func NewREPLInteractorParams(cfg *config.Config, env *entity.Env) (*usecase.InteractorParams, error) {
	return newInteractorParams(cfg, env, gateway.NewPrompt(cfg, env))
}

func newInteractorParams(cfg *config.Config, env *entity.Env, inputter port.Inputter) (*usecase.InteractorParams, error) {
	var result error
	grpcAdapter, err := gateway.NewGRPCClient(cfg)
	if err != nil {
		result = multierror.Append(result, err)
	}
	return &usecase.InteractorParams{
		Env:            env,
		OutputPort:     presenter.NewJSONCLIPresenterWithIndent(),
		InputterPort:   inputter,
		GRPCClient:     grpcAdapter,
		DynamicBuilder: gateway.NewDynamicBuilder(),
	}, result
}
