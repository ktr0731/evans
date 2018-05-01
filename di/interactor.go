package di

import (
	"sync"

	"github.com/ktr0731/evans/adapter/controller"
	"github.com/ktr0731/evans/adapter/gateway"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/pkg/errors"
)

func NewCLIInteractor(cfg *config.Config, opt *controller.Options, env *entity.Env) (*usecase.Interactor, error) {
	grpcAdapter, err := gRPCAdapter(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CLI interactor")
	}
	param := &usecase.InteractorParams{
		Env:            env,
		OutputPort:     presenter.NewJSONCLIPresenterWithIndent(),
		GRPCPort:       grpcAdapter,
		DynamicBuilder: gateway.NewDynamicBuilder(),
	}
	return usecase.NewInteractor(param), nil
}

func NewREPLInteractor() *usecase.Interactor {
	return nil
}

var grpcOnce sync.Once

func gRPCAdapter(cfg *config.Config) (port.GRPCPort, error) {
	return gateway.NewGRPCClient(cfg)
}
