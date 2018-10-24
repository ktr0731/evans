// +build wireinject
package di

import (
	"io"

	"github.com/ktr0731/evans/adapter/gateway"
	"github.com/ktr0731/evans/adapter/grpc"
	"github.com/ktr0731/evans/adapter/parser"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	environment "github.com/ktr0731/evans/entity/env"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/pkg/errors"
)

func provideEnv(cfg *config.Config) (environment.Environment, error) {
	paths, err := resolveProtoPaths(cfg)
	if err != nil {
		return nil, err
	}

	files := resolveProtoFiles(cfg)
	desc, err := parser.ParseFile(files, paths)
	if err != nil {
		return nil, err
	}

	var svcs []entity.Service
	gRPCClient, err := GRPCClient(cfg)
	if err != nil {
		return nil, err
	}

	var env environment.Environment

	if gRPCClient.ReflectionEnabled() {
		svcs, err = gRPCClient.ListServices()
		if err != nil {
			return nil, errors.Wrap(err, "failed to list services by gRPC reflection")
		}
		env = environment.NewFromServices(svcs, cfg)
	} else {
		env = environment.New(desc, cfg)

		if pkg := cfg.Default.Package; pkg != "" {
			if err := env.UsePackage(pkg); err != nil {
				return nil, errors.Wrapf(err, "failed to set package to env as a default package: %s", pkg)
			}
		}
	}

	if svc := cfg.Default.Service; svc != "" {
		if err := env.UseService(svc); err != nil {
			return nil, errors.Wrapf(err, "failed to set service to env as a default service: %s", svc)
		}
	}

	return env, nil
}

func provideJSONPresenter() *presenter.JSONPresenter {
	return presenter.NewJSONWithIndent()
}

type JSONReader io.Reader

func provideJSONReader() JSONReader {
	// TODO:
	return nil
}

func provideJSONInputter(in JSONReader) *gateway.JSONFileInputter {
	return gateway.NewJSONFileInputter(in)
}

func providePromptInputter(cfg *config.Config, env environment.Environment) *gateway.Prompt {
	return gateway.NewPrompt(cfg, env)
}

func provideGRPCClient(cfg *config.Config, builder port.DynamicBuilder) (entity.GRPCClient, error) {
	var client entity.GRPCClient
	if cfg.Request.Web {
		client = grpc.NewWebClient(cfg, builder)
	} else {
		var err error
		client, err = grpc.NewClient(cfg)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

func provideDynamicBuilder() *gateway.DynamicBuilder {
	return gateway.NewDynamicBuilder()
}
