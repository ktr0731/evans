package config

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ktr0731/evans/logger"
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

var (
	update = flag.Bool("update", false, "update golden files")
)

func init() {
	if *update {
		os.RemoveAll(filepath.Join("testdata", "fixtures"))
	}
}

// setupEnv creates a temp dir and set $XDG_CONFIG_HOME to it.
// setupEnv returns dir which is the base dir and cleanup func.
//
// The directory structure is as follows:
//
//   - (temp dir): dir
//     - config: $XDG_CONFIG_HOME
//       - evans: evansCfgDir
//
func setupEnv(t *testing.T) (string, string, func()) {
	cwd := getWorkDir(t)

	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, "failed to create a temp dir to setup testing enviroment")

	os.Chdir(dir)
	oldEnv := os.Getenv("XDG_CONFIG_HOME")
	cfgDir := filepath.Join(dir, "config")
	os.Setenv("XDG_CONFIG_HOME", cfgDir)
	evansCfgDir := filepath.Join(cfgDir, "evans")
	mkdir(t, evansCfgDir)

	return dir, evansCfgDir, func() {
		os.Chdir(cwd)
		os.Setenv("XDG_CONFIG_HOME", oldEnv)
		os.RemoveAll(dir)
		viper.Reset()
	}
}

func assertWithGolden(t *testing.T, name string, f func(t *testing.T) *Config) {
	normalizeFilename := func(name string) string {
		fname := strings.Replace(strings.ToLower(name), " ", "_", -1) + ".golden.toml"
		return filepath.Join("testdata", "fixtures", fname)
	}

	t.Run(name, func(t *testing.T) {
		t.Helper()
		cfg := f(t)

		fname := normalizeFilename(name)

		// Load a TOML formatted golden file.
		v := viper.New()
		v.SetConfigType("toml")
		f, err := os.Open(fname)
		if *update {
			createGolden(t, fname, cfg)
			logger.Printf("golden updated: %s", fname)
			return
		}
		require.NoError(t, err, "failed to load a golden file")
		defer f.Close()

		err = v.ReadConfig(f)
		require.NoError(t, err, "failed to read golden file")

		var expected Config
		err = v.Unmarshal(&expected)
		require.NoError(t, err, "failed to unmarshal a golden file")
		setupConfig(&expected)

		assert.EqualValues(t, expected, *cfg)
	})
}

func createGolden(t *testing.T, fname string, cfg *Config) {
	t.Helper()

	m := structToMap(cfg).(map[string]interface{})
	tree, err := toml.TreeFromMap(m)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(fname)
	require.NoError(t, err, "failed to create a file")
	defer f.Close()
	_, err = tree.WriteTo(f)
	require.NoError(t, err, "failed to encode cfg as a TOML format")
}

func structToMap(i interface{}) interface{} {
	rv := reflect.ValueOf(i)
	if rv.Kind() != reflect.Struct && rv.Kind() != reflect.Ptr {
		return rv.Interface()
	}

	m := make(map[string]interface{})
	el := rv.Elem()
	for i := 0; i < el.Type().NumField(); i++ {
		// spf13/viper formats all keys to lower-case.
		tag := strings.ToLower(el.Type().Field(i).Tag.Get("toml"))
		iface := el.Field(i).Interface()
		m[tag] = structToMap(iface)
	}
	return m
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestLoad(t *testing.T) {
	checkValues := func(t *testing.T, c *Config) {
		require.NotNil(t, c.REPL.Server)
		if len(c.Default.ProtoFile) == 1 {
			require.NotEmpty(t, c.Default.ProtoFile[0])
		}
		if len(c.Default.ProtoPath) == 1 {
			require.NotEmpty(t, c.Default.ProtoPath[0])
		}
	}

	assertWithGolden(t, "create a default global config if both of global and local config are not found", func(t *testing.T) *Config {
		_, _, cleanup := setupEnv(t)
		defer cleanup()

		cfg, err := Get(nil)
		require.NoError(t, err, "Get must not return any errors")

		checkValues(t, cfg)

		return cfg
	})

	assertWithGolden(t, "load a global config if local config is not found", func(t *testing.T) *Config {
		oldCWD := getWorkDir(t)

		_, cfgDir, cleanup := setupEnv(t)
		defer cleanup()

		// Copy global.toml from testdata to the config dir.
		// default config was changed host and port to 'localhost' and '3000'.
		copyFile(t, filepath.Join(cfgDir, "config.toml"), filepath.Join(oldCWD, "testdata", "global.toml"))

		cfg, err := Get(nil)
		require.NoError(t, err, "Get must not return any errors")

		checkValues(t, cfg)

		return cfg
	})

	assertWithGolden(t, "load a local config", func(t *testing.T) *Config {
		oldCWD := getWorkDir(t)

		cwd, cfgDir, cleanup := setupEnv(t)
		defer cleanup()

		projDir := filepath.Join(cwd, "local")
		mkdir(t, projDir)

		// Copy global.toml from testdata to the config dir.
		// default config was changed host and port to 'localhost' and '3000'.
		copyFile(t, filepath.Join(cfgDir, "config.toml"), filepath.Join(oldCWD, "testdata", "global.toml"))
		// Copy local.toml from testdata to the project dir.
		// global.toml was changed protopath, request.header and port to '["bar"]', 'grpc-client = "evans"' and '3333'.
		copyFile(t, filepath.Join(projDir, ".evans.toml"), filepath.Join(oldCWD, "testdata", "local.toml"))

		os.Chdir(projDir)
		err := exec.Command("git", "init").Run()
		require.NoError(t, err, "failed to init a pseudo project")

		cfg, err := Get(nil)
		require.NoError(t, err, "Get must not return any errors")

		checkValues(t, cfg)

		return cfg
	})

	assertWithGolden(t, "will be apply flags", func(t *testing.T) *Config {
		oldCWD := getWorkDir(t)

		cwd, cfgDir, cleanup := setupEnv(t)
		defer cleanup()

		projDir := filepath.Join(cwd, "local")

		mkdir(t, projDir)
		// Copy global.toml from testdata to the config dir.
		// default config was changed host and port to 'localhost' and '3000'.
		copyFile(t, filepath.Join(cfgDir, "config.toml"), filepath.Join(oldCWD, "testdata", "global.toml"))
		// Copy local.toml from testdata to the project dir.
		// global.toml was changed protopath, request.header and port to '["bar"]', 'grpc-client = "evans"' and '3333'.
		copyFile(t, filepath.Join(projDir, ".evans.toml"), filepath.Join(oldCWD, "testdata", "local.toml"))

		os.Chdir(projDir)
		err := exec.Command("git", "init").Run()
		require.NoError(t, err, "failed to init a pseudo project")

		fs := pflag.NewFlagSet("test", pflag.ExitOnError)
		fs.String("port", "", "")
		fs.StringToString("header", nil, "")
		fs.StringSlice("path", nil, "")
		// --port flag changes port number to '8080'. Also --header appends 'foo=bar' and 'hoge=fuga' to 'request.header'.
		fs.Parse([]string{
			"--port", "8080",
			"--path", "yoko.touma",
			"--header", "foo=bar", "--header", "hoge=fuga",
		})

		cfg, err := Get(fs)
		require.NoError(t, err, "Get must not return any errors")

		checkValues(t, cfg)

		return cfg
	})

	assertWithGolden(t, "will be apply flags, but local config doesn't exist", func(t *testing.T) *Config {
		oldCWD := getWorkDir(t)

		_, cfgDir, cleanup := setupEnv(t)
		defer cleanup()

		// Copy global.toml from testdata to the config dir.
		// default config was changed host and port to 'localhost' and '3000'.
		copyFile(t, filepath.Join(cfgDir, "config.toml"), filepath.Join(oldCWD, "testdata", "global.toml"))

		fs := pflag.NewFlagSet("test", pflag.ExitOnError)
		fs.String("port", "", "")
		// --port flag changes port number to '8080'.
		fs.Parse([]string{"--port", "8080"})

		cfg, err := Get(fs)
		require.NoError(t, err, "Get must not return any errors")

		checkValues(t, cfg)

		return cfg
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

func mkdir(t *testing.T, dir string) {
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err, "failed to create dirs")
}
