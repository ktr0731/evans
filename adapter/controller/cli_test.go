package controller

import (
	"os"
	"testing"

	"github.com/ktr0731/evans/config"
	"github.com/stretchr/testify/require"
)

func TestCLI(t *testing.T) {
	newConfig := func() *config.Config {
		return &config.Config{
			Default: &config.Default{},
		}
	}

	t.Run("proto path (option)", func(t *testing.T) {
		conf := newConfig()
		opt := &options{
			path: []string{"foo", "foo", "bar"},
		}
		paths, err := collectProtoPaths(conf, opt, nil)
		require.NoError(t, err)
		require.Len(t, paths, 2)
	})

	t.Run("proto path (config)", func(t *testing.T) {
		conf := newConfig()
		conf.Default.ProtoPath = []string{"foo", "foo", "bar"}
		opt := &options{}
		paths, err := collectProtoPaths(conf, opt, nil)
		require.NoError(t, err)
		require.Len(t, paths, 2)
	})

	t.Run("proto path (both)", func(t *testing.T) {
		conf := newConfig()
		conf.Default.ProtoPath = []string{"foo", "foo", "bar"}
		opt := &options{
			path: []string{"foo", "baz", "bar"},
		}
		paths, err := collectProtoPaths(conf, opt, nil)
		require.NoError(t, err)
		require.Len(t, paths, 3)
	})

	t.Run("proto path (from file)", func(t *testing.T) {
		conf := newConfig()
		proto := []string{"./hoge", "./foo/bar"}
		opt := &options{}
		paths, err := collectProtoPaths(conf, opt, proto)
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
		conf := newConfig()
		proto := []string{"$hoge/foo", "/fuga/bar"}
		opt := &options{}

		cleanup := setEnv("hoge", "/fuga")
		defer cleanup()

		paths, err := collectProtoPaths(conf, opt, proto)
		require.NoError(t, err)
		require.Len(t, paths, 1)
		require.Equal(t, "/fuga", paths[0])
	})

	t.Run("error/proto path", func(t *testing.T) {
		conf := newConfig()
		proto := []string{"foo bar"}
		opt := &options{}

		_, err := collectProtoPaths(conf, opt, proto)
		require.Error(t, err)
	})
}
