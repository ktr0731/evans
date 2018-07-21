package di

import (
	"errors"
	"io"
	"path/filepath"
	"sync"

	"github.com/ktr0731/evans/adapter/gateway"
	"github.com/ktr0731/evans/adapter/parser"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase"
	multierror "github.com/ktr0731/go-multierror"
	shellwords "github.com/mattn/go-shellwords"
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
	})
	return
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
	jsonFIleInputterOnce sync.Once
)

func initJSONFileInputter(in io.Reader) error {
	jsonFIleInputterOnce.Do(func() {
		jsonFileInputter = gateway.NewJSONFileInputter(in)
	})
	return nil
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

var (
	dynamicBuilder     *gateway.DynamicBuilder
	dynamicBuilderOnce sync.Once
)

func initDynamicBuilder() error {
	dynamicBuilder = gateway.NewDynamicBuilder()
	return nil
}

var (
	initOnce sync.Once
)

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
			func() error { return initGRPCClient(cfg) },
			func() error { return initEnv(cfg) },
			initJSONCLIPresenter,
			initDynamicBuilder,
		)
	})
	return initer.init()
}

func NewCLIInteractorParams(cfg *config.Config, in io.Reader) (*usecase.InteractorParams, error) {
	if err := initDependencies(cfg, in); err != nil {
		return nil, err
	}
	return &usecase.InteractorParams{
		Env:            env,
		OutputPort:     jsonCLIPresenter,
		InputterPort:   jsonFileInputter,
		GRPCClient:     gRPCClient,
		DynamicBuilder: dynamicBuilder,
	}, nil
}

func NewREPLInteractorParams(cfg *config.Config, in io.Reader) (param *usecase.InteractorParams, err error) {
	if err := initDependencies(cfg, in); err != nil {
		return nil, err
	}
	return &usecase.InteractorParams{
		Env:            env,
		OutputPort:     jsonCLIPresenter,
		InputterPort:   jsonFileInputter,
		GRPCClient:     gRPCClient,
		DynamicBuilder: dynamicBuilder,
	}, nil
}
