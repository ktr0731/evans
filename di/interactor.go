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

func NewCLIInteractor(cfg *config.Config, env *entity.Env, inputter port.Inputter) (*usecase.Interactor, error) {
	return newInteractor(cfg, env, inputter)
}

func NewREPLInteractor(cfg *config.Config, env *entity.Env) (*usecase.Interactor, error) {
	inputter := gateway.NewPrompt(cfg, env)
	return newInteractor(cfg, env, inputter)
}

func newInteractor(cfg *config.Config, env *entity.Env, inputter port.Inputter) (*usecase.Interactor, error) {
	var result error
	grpcAdapter, err := gateway.NewGRPCClient(cfg)
	if err != nil {
		result = multierror.Append(result, err)
	}
	return usecase.NewInteractor(&usecase.InteractorParams{
		Env:            env,
		OutputPort:     presenter.NewJSONCLIPresenterWithIndent(),
		InputterPort:   inputter,
		GRPCPort:       grpcAdapter,
		DynamicBuilder: gateway.NewDynamicBuilder(),
	}), result
}
