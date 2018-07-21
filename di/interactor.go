package di

import (
	"sync"

	"github.com/ktr0731/evans/adapter/gateway"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/evans/usecase/port"
	multierror "github.com/ktr0731/go-multierror"
)

var (
	env     *entity.Env
	envOnce sync.Once
)

func injectEnv(cfg *config.Config) *entity.Env {
	envOnce.Do(func() {
		env = entity.NewEnv()
	})
	return nil
}

func NewCLIInteractorParams(cfg *config.Config, env *entity.Env, inputter port.Inputter) (*usecase.InteractorParams, error) {
	return &usecase.InteractorParams{
		Env:            injectEnv(cfg),
		OutputPort:     injectJSONCLIPresenterWithIndent(),
		InputterPort:   injectJSONFileInputter(),
		GRPCClient:     injectGRPCClient(),
		DynamicBuilder: injectNewDynamicBuilder(),
	}
}

func NewREPLInteractorParams(cfg *config.Config, env *entity.Env) (*usecase.InteractorParams, error) {
	return &usecase.InteractorParams{
		Env:            injectEnv(cfg),
		OutputPort:     injectJSONCLIPresenterWithIndent(),
		InputterPort:   injectJSONFileInputter(),
		GRPCClient:     injectGRPCClient(),
		DynamicBuilder: injectDynamicBuilder(),
	}
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
