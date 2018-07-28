package di

import (
	"io"
	"path/filepath"
	"sync"

	"github.com/k0kubun/pp"
	"github.com/ktr0731/evans/adapter/gateway"
	"github.com/ktr0731/evans/adapter/parser"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	multierror "github.com/ktr0731/go-multierror"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
)

var (
	env     *entity.Env
	envOnce sync.Once
)

func initEnv(cfg *config.Config) (rerr error) {
	envOnce.Do(func() {
		paths, err := resolveProtoPaths(cfg)
		if err != nil {
			rerr = err
			return
		}

		files := resolveProtoFiles(cfg)
		desc, err := parser.ParseFile(files, paths)
		if err != nil {
			rerr = err
			return
		}

		env = entity.NewEnv(desc, cfg)

		if pkg := cfg.Default.Package; pkg != "" {
			if err := env.UsePackage(pkg); err != nil {
				rerr = errors.Wrapf(err, "failed to set package to env as a default package: %s", pkg)
				return
			}
		}

		if svc := cfg.Default.Service; svc != "" {
			if err := env.UseService(svc); err != nil {
				rerr = errors.Wrapf(err, "failed to set service to env as a default service: %s", svc)
				return
			}
		}

		gRPCClient, err := GRPCClient(cfg)
		if err != nil {
			rerr = err
			return
		}

		if gRPCClient.ReflectionEnabled() {
			for _, s := range gRPCClient.ListServices() {
				pp.Println(s.FQRN())
			}
		}
	})
	return
}

func Env(cfg *config.Config) (entity.Environment, error) {
	if err := initEnv(cfg); err != nil {
		return nil, err
	}
	return env, nil
}

func resolveProtoPaths(cfg *config.Config) ([]string, error) {
	paths := make([]string, 0, len(cfg.Default.ProtoPath))
	encountered := map[string]bool{}
	parser := shellwords.NewParser()
	parser.ParseEnv = true

	parse := func(p string) (string, error) {
		res, err := parser.Parse(p)
		if err != nil {
			return "", err
		}
		if len(res) > 1 {
			return "", errors.New("failed to parse proto path")
		}
		// empty path
		if len(res) == 0 {
			return "", nil
		}
		return res[0], nil
	}

	fpaths := make([]string, 0, len(cfg.Default.ProtoFile))
	for _, f := range cfg.Default.ProtoFile {
		fpaths = append(fpaths, filepath.Dir(f))
	}

	for _, p := range append(cfg.Default.ProtoPath, fpaths...) {
		path, err := parse(p)
		if err != nil {
			return nil, err
		}

		if encountered[path] || path == "" {
			continue
		}
		encountered[path] = true
		paths = append(paths, path)
	}

	return paths, nil
}

func resolveProtoFiles(conf *config.Config) []string {
	files := make([]string, 0, len(conf.Default.ProtoFile))
	for _, f := range conf.Default.ProtoFile {
		if f != "" {
			files = append(files, f)
		}
	}
	return files
}

var (
	jsonCLIPresenter     *presenter.CLIPresenter
	jsonCLIPresenterOnce sync.Once
)

func initJSONCLIPresenter() error {
	jsonCLIPresenterOnce.Do(func() {
		jsonCLIPresenter = presenter.NewJSONCLIPresenterWithIndent()
	})
	return nil
}

var (
	jsonFileInputter     *gateway.JSONFileInputter
	jsonFileInputterOnce sync.Once
)

func initJSONFileInputter(in io.Reader) error {
	jsonFileInputterOnce.Do(func() {
		jsonFileInputter = gateway.NewJSONFileInputter(in)
	})
	return nil
}

var (
	promptInputter     *gateway.Prompt
	promptInputterOnce sync.Once
)

func initPromptInputter(cfg *config.Config) (err error) {
	promptInputterOnce.Do(func() {
		var e entity.Environment
		e, err = Env(cfg)
		promptInputter = gateway.NewPrompt(cfg, e)
	})
	return
}

var (
	gRPCClient     *gateway.GRPCClient
	gRPCClientOnce sync.Once
)

func initGRPCClient(cfg *config.Config) error {
	var err error
	gRPCClientOnce.Do(func() {
		gRPCClient, err = gateway.NewGRPCClient(cfg)
	})
	return err
}

func GRPCClient(cfg *config.Config) (*gateway.GRPCClient, error) {
	if err := initGRPCClient(cfg); err != nil {
		return nil, err
	}
	return gRPCClient, nil
}

var (
	dynamicBuilder     *gateway.DynamicBuilder
	dynamicBuilderOnce sync.Once
)

func initDynamicBuilder() error {
	dynamicBuilderOnce.Do(func() {
		dynamicBuilder = gateway.NewDynamicBuilder()
	})
	return nil
}

type initializer struct {
	f []func() error

	resultCache error
	done        bool
}

func (i *initializer) register(f ...func() error) {
	i.f = append(i.f, f...)
}

func (i *initializer) init() error {
	if i.done {
		return i.resultCache
	}

	i.done = true

	var result error
	for _, f := range i.f {
		if err := f(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

var (
	initer     *initializer
	initerOnce sync.Once
)

func initDependencies(cfg *config.Config, in io.Reader) error {
	initerOnce.Do(func() {
		initer = &initializer{}
		initer.register(
			func() error { return initJSONFileInputter(in) },
			func() error { return initPromptInputter(cfg) },
			func() error { return initGRPCClient(cfg) },
			func() error { return initEnv(cfg) },
			initJSONCLIPresenter,
			initDynamicBuilder,
		)
	})
	return initer.init()
}
