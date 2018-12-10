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

func assertWithGolden(t *testing.T, name string, f func(t *testing.T) *Config) {
	normalizeFilename := func(name string) string {
		fname := strings.Replace(strings.ToLower(name), " ", "_", -1) + ".golden.toml"
		return filepath.Join("testdata", "fixtures", fname)
	}

	t.Run(name, func(t *testing.T) {
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
	setup := func(t *testing.T) (string, func()) {
		cwd := getWorkDir(t)

		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err, "failed to create a temp dir to setup testing enviroment")

		os.Chdir(dir)
		oldEnv := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

		return dir, func() {
			os.Chdir(cwd)
			os.Setenv("XDG_CONFIG_HOME", oldEnv)
			os.RemoveAll(dir)
			viper.Reset()
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

	assertWithGolden(t, "create a default global config if both of global and local config are not found", func(t *testing.T) *Config {
		_, cleanup := setup(t)
		defer cleanup()

		cfg, err := Get(nil)
		require.NoError(t, err, "Get must not return any errors")

		checkValues(t, cfg)

		return cfg
	})

	assertWithGolden(t, "load a global config if local config is not found", func(t *testing.T) *Config {
		oldCWD := getWorkDir(t)

		cwd, cleanup := setup(t)
		defer cleanup()

		err := os.MkdirAll(filepath.Join(cwd, "config", "evans"), 0755)
		require.NoError(t, err, "failed to setup config dir")
		// Copy global.toml from testdata to the config dir.
		// global.toml was changed host and port to 'localhost' and '3000'.
		copyFile(t, filepath.Join(cwd, "config", "evans", "config.toml"), filepath.Join(oldCWD, "testdata", "global.toml"))

		cfg, err := Get(nil)
		require.NoError(t, err, "Get must not return any errors")

		checkValues(t, cfg)

		return cfg
	})

	assertWithGolden(t, "load a local config", func(t *testing.T) *Config {
		oldCWD := getWorkDir(t)

		cwd, cleanup := setup(t)
		defer cleanup()

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

		return cfg
	})

	assertWithGolden(t, "will be apply flags", func(t *testing.T) *Config {
		oldCWD := getWorkDir(t)

		cwd, cleanup := setup(t)
		defer cleanup()

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
