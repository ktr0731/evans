package controller

import (
	"os"
	"testing"

	"github.com/ktr0731/evans/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI(t *testing.T) {
	newConfig := func() *config.Config {
		return &config.Config{
			Default: &config.Default{},
		}
	}

	t.Run("proto path (option)", func(t *testing.T) {
		cfg := newConfig()
		cfg.Default.ProtoPath = []string{"foo", "foo", "bar"}
		paths, err := resolveProtoPaths(cfg)
		require.NoError(t, err)
		require.Len(t, paths, 2)
	})

	t.Run("proto path (config)", func(t *testing.T) {
		cfg := newConfig()
		cfg.Default.ProtoPath = []string{"foo", "foo", "bar"}
		paths, err := resolveProtoPaths(cfg)
		require.NoError(t, err)
		require.Len(t, paths, 2)
	})

	t.Run("proto path (from file)", func(t *testing.T) {
		cfg := newConfig()
		cfg.Default.ProtoFile = []string{"./hoge", "./foo/bar"}
		paths, err := resolveProtoPaths(cfg)
		require.NoError(t, err)
		require.Len(t, paths, 2)
		require.Exactly(t, []string{".", "foo"}, paths)
	})

	setEnv := func(k, v string) func() {
		old := os.Getenv(k)
		os.Setenv(k, v)
		return func() {
			os.Setenv(k, old)
		}
	}

	t.Run("proto path (env)", func(t *testing.T) {
		cfg := newConfig()
		cfg.Default.ProtoFile = []string{"$hoge/foo", "/fuga/bar"}

		cleanup := setEnv("hoge", "/fuga")
		defer cleanup()

		paths, err := resolveProtoPaths(cfg)
		require.NoError(t, err)
		require.Len(t, paths, 1)
		require.Equal(t, "/fuga", paths[0])
	})

	t.Run("error/proto path", func(t *testing.T) {
		cfg := newConfig()
		cfg.Default.ProtoPath = []string{"foo bar"}

		_, err := resolveProtoPaths(cfg)
		require.Error(t, err)
	})
}

func Test_mergeConfig(t *testing.T) {
	// setup
	// append elements to slice which will be merged.
	cfg := &config.Config{
		Default: &config.Default{
			Package:   "tamaki",
			ProtoPath: []string{"kobuchizawa"},
			ProtoFile: []string{"miyake"},
		},
		Server: &config.Server{
			Port: "50052",
		},
		Request: &config.Request{
			Header: []config.Header{
				{Key: "yuzuki", Val: "shiraishi"},
				{Key: "nozomi", Val: "kasaki"},
				{Key: "nozomi", Val: "kasaki2"},
			},
		},
		Env:  &config.Env{},
		REPL: &config.REPL{},
	}
	config.SetupConfig(cfg)

	opt := &options{
		pkg:     "kumiko",
		service: "reina",
		path:    []string{"kobuchizawa", "midori"},
		host:    "hazuki",
		port:    "50053",
		header:  []string{"nozomi=kasaki"},
	}
	proto := []string{"noboru"}

	res, err := mergeConfig(cfg, opt, proto)
	require.NoError(t, err)
	assert.Equal(t, res.REPL.Server, res.Server)
	assert.Equal(t, res.Env.Server, res.Server)
	assert.Equal(t, opt.pkg, res.Default.Package)
	assert.Equal(t, opt.service, res.Default.Service)
	assert.Equal(t, []string(opt.path), res.Default.ProtoPath)
	assert.Equal(t, append(cfg.Default.ProtoFile, proto...), res.Default.ProtoFile)
	assert.Equal(t, opt.host, res.Server.Host)
	assert.Equal(t, opt.port, res.Server.Port)
	assert.Equal(t, cfg.Request.Header[:2], res.Request.Header)
}
