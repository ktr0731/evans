package config

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestLoad(t *testing.T) {
	setup := func(t *testing.T) (string, func()) {
		cwd := getWorkDir(t)

		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err, "failed to create a temp dir to setup testing enviroment")

		os.Chdir(dir)

		return dir, func() {
			os.Chdir(cwd)
			os.RemoveAll(dir)
			viper.Reset()
		}
	}

	changeEnv := func(k, v string) func() {
		old := os.Getenv(k)
		os.Setenv(k, v)
		return func() {
			os.Setenv(k, old)
		}
	}

	checkValues := func(t *testing.T, c *Config) {
		require.NotNil(t, c.REPL.Server)
		if len(c.Default.ProtoFile) == 1 {
			require.NotEmpty(t, c.Default.ProtoFile[0])
		}
		if len(c.Default.ProtoPath) == 1 {
			require.NotEmpty(t, c.Default.ProtoPath[0])
		}
	}

	t.Run("Create a default global config if both of global and local config are not found", func(t *testing.T) {
		cwd, cleanup := setup(t)
		defer cleanup()
		resetEnv := changeEnv("XDG_CONFIG_HOME", filepath.Join(cwd, "config"))
		defer resetEnv()

		cfg, err := Get(nil)
		require.NoError(t, err, "Get must not return any errors")

		checkValues(t, cfg)

		defCfg := DefaultConfig(t)
		assert.EqualValues(t, cfg, defCfg)
	})

	t.Run("Load a global config if local config are not found", func(t *testing.T) {
		oldCWD := getWorkDir(t)

		cwd, cleanup := setup(t)
		defer cleanup()
		resetEnv := changeEnv("XDG_CONFIG_HOME", filepath.Join(cwd, "config"))
		defer resetEnv()

		err := os.MkdirAll(filepath.Join(cwd, "config", "evans"), 0755)
		require.NoError(t, err, "failed to setup config dir")
		// Copy global.toml from testdata to the config dir.
		// global.toml was changed host and port to 'localhost' and '3000'.
		copyFile(t, filepath.Join(cwd, "config", "evans", "config.toml"), filepath.Join(oldCWD, "testdata", "global.toml"))

		cfg, err := Get(nil)
		require.NoError(t, err, "Get must not return any errors")

		checkValues(t, cfg)

		expected := DefaultConfig(t)
		expected.Server.Host = "localhost"
		expected.Server.Port = "3000"
		expected.Default.ProtoPath = []string{"foo"}

		assert.EqualValues(t, expected, cfg)
	})

	t.Run("Will be apply local config if global/local config files are found", func(t *testing.T) {
		oldCWD := getWorkDir(t)

		cwd, cleanup := setup(t)
		defer cleanup()
		resetEnv := changeEnv("XDG_CONFIG_HOME", filepath.Join(cwd, "config"))
		defer resetEnv()

		projDir := filepath.Join(cwd, "local")

		err := os.MkdirAll(filepath.Join(cwd, "config", "evans"), 0755)
		require.NoError(t, err, "failed to setup config dir")
		err = os.MkdirAll(projDir, 0755)
		require.NoError(t, err, "failed to setup a project local dir")
		copyFile(t, filepath.Join(cwd, "config", "evans", "config.toml"), filepath.Join(oldCWD, "testdata", "global.toml"))
		copyFile(t, filepath.Join(projDir, ".evans.toml"), filepath.Join(oldCWD, "testdata", "local.toml"))

		os.Chdir(projDir)
		err = exec.Command("git", "init").Run()
		require.NoError(t, err, "failed to init a pseudo project")

		cfg, err := Get(nil)
		require.NoError(t, err, "Get must not return any errors")

		checkValues(t, cfg)

		expected := DefaultConfig(t)
		// Global config
		expected.Server.Host = "localhost"
		// Local config (global config is overwritten)
		expected.Server.Port = "3333"
		expected.Request.Header["foo"] = "bar"
		expected.Default.ProtoPath = []string{"bar"}

		assert.EqualValues(t, expected, cfg)
	})

	t.Run("Will be apply local config and flags", func(t *testing.T) {
		oldCWD := getWorkDir(t)

		cwd, cleanup := setup(t)
		defer cleanup()
		resetEnv := changeEnv("XDG_CONFIG_HOME", filepath.Join(cwd, "config"))
		defer resetEnv()

		projDir := filepath.Join(cwd, "local")

		err := os.MkdirAll(filepath.Join(cwd, "config", "evans"), 0755)
		require.NoError(t, err, "failed to setup config dir")
		err = os.MkdirAll(projDir, 0755)
		require.NoError(t, err, "failed to setup a project local dir")
		copyFile(t, filepath.Join(cwd, "config", "evans", "config.toml"), filepath.Join(oldCWD, "testdata", "global.toml"))
		copyFile(t, filepath.Join(projDir, ".evans.toml"), filepath.Join(oldCWD, "testdata", "local.toml"))

		os.Chdir(projDir)
		err = exec.Command("git", "init").Run()
		require.NoError(t, err, "failed to init a pseudo project")

		fs := pflag.NewFlagSet("test", pflag.ExitOnError)
		fs.String("port", "", "")
		fs.Parse([]string{"--port", "8080"})

		cfg, err := Get(fs)
		require.NoError(t, err, "Get must not return any errors")

		checkValues(t, cfg)

		expected := DefaultConfig(t)
		// Global config
		expected.Server.Host = "localhost"
		// Local config (global config is overwritten)
		expected.Request.Header["foo"] = "bar"
		expected.Default.ProtoPath = []string{"bar"}
		// Flags (global and local configs are overwritten)
		expected.Server.Port = "8080"

		assert.EqualValues(t, expected, cfg)
	})
}

func getWorkDir(t *testing.T) string {
	cwd, err := os.Getwd()
	require.NoError(t, err, "failed to get the working dir")
	return cwd
}

func copyFile(t *testing.T, to, from string) {
	tf, err := os.Create(to)
	require.NoError(t, err, "failed to create a config file")
	defer tf.Close()
	ff, err := os.Open(from)
	require.NoError(t, err, "failed to open a prepared config file")
	defer ff.Close()
	io.Copy(tf, ff)
}
