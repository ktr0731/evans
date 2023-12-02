package config

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/logger"
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/goleak"
)

var (
	update = flag.Bool("update", false, "update golden files")
)

func init() {
	testing.Init()
	flag.Parse()
	if *update {
		os.RemoveAll(filepath.Join("testdata", "fixtures"))
		err := os.Mkdir(filepath.Join("testdata", "fixtures"), 0744)
		if err != nil {
			panic(fmt.Sprintf("failed to create fixtures dir: %s", err))
		}
	}
}

// setupEnv creates a temp dir and set $XDG_CONFIG_HOME to it.
// setupEnv returns dir which is the base dir and cleanup func.
//
// The directory structure is as follows:
//
//	─ (temp dir): dir
//	   ─ config: $XDG_CONFIG_HOME
//	      ─ evans: evansCfgDir
func setupEnv(t *testing.T) (string, string, func()) {
	cwd := getWorkDir(t)

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("failed to create a temp dir to setup testing environment: %s", err)
	}

	mustChdir(t, dir)
	cfgDir := filepath.Join(dir, "config")
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	evansCfgDir := filepath.Join(cfgDir, "evans")
	mkdir(t, evansCfgDir)

	return dir, evansCfgDir, func() {
		mustChdir(t, cwd)
		os.RemoveAll(dir)
		viper.Reset()
	}
}

func assertWithGolden(t *testing.T, name string, f func(t *testing.T) *Config) {
	t.Helper()

	r := strings.NewReplacer(
		" ", "_",
		"=", "-",
		"'", "",
		`"`, "",
		",", "",
	)
	normalizeFilename := func(name string) string {
		fname := r.Replace(strings.ToLower(name)) + ".golden.toml"
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
		if err != nil {
			t.Fatalf("failed to load a golden file: %s", err)
		}
		defer f.Close()

		err = v.ReadConfig(f)
		if err != nil {
			t.Fatalf("failed to read golden file: %s", err)
		}

		var expected Config
		err = v.Unmarshal(&expected)
		if err != nil {
			t.Fatalf("failed to unmarshal a golden file: %s", err)
		}
		setupConfig(&expected)

		if diff := cmp.Diff(expected, *cfg); diff != "" {
			t.Errorf("-want, +got\n%s", diff)
		}
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
	if err != nil {
		t.Fatalf("failed to create a file: %s", err)
	}
	defer f.Close()
	_, err = tree.WriteTo(f)
	if err != nil {
		t.Fatalf("failed to encode cfg as a TOML format: %s", err)
	}
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
		if len(c.Default.ProtoFile) >= 1 {
			if len(c.Default.ProtoFile[0]) == 0 {
				t.Fatalf("Default.ProtoFile must not be empty")
			}
		}
		if len(c.Default.ProtoPath) >= 1 {
			if len(c.Default.ProtoPath[0]) == 0 {
				t.Fatalf("Default.ProtoPath must not be empty")
			}
		}
	}
	assertWithGolden(t, "create a default global config if both of global and local config are not found",
		func(t *testing.T) *Config {
			_, _, cleanup := setupEnv(t)
			defer cleanup()

			cfg := mustGet(t, nil)

			checkValues(t, cfg)

			return cfg
		})

	t.Run("load an environmental variable that overrides local config", func(t *testing.T) {
		t.Setenv("EVANS_SERVER_PORT", "9001")

		_, _, cleanup := setupEnv(t)
		defer cleanup()

		cfg := mustGet(t, nil)

		if cfg.Server.Port != "9001" {
			t.Errorf("port %s not set by os env", cfg.Server.Port)
		}
	})

	assertWithGolden(t, "load a global config if local config is not found", func(t *testing.T) *Config {
		oldCWD := getWorkDir(t)

		_, cfgDir, cleanup := setupEnv(t)
		defer cleanup()

		// Copy global.toml from testdata to the config dir.
		// default config was changed host and port to 'localhost' and '3000'.
		copyFile(t, filepath.Join(cfgDir, "config.toml"), filepath.Join(oldCWD, "testdata", "global.toml"))

		cfg := mustGet(t, nil)

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

		mustChdir(t, projDir)
		err := exec.Command("git", "init").Run()
		if err != nil {
			t.Fatalf("failed to init a pseudo project: %s", err)
		}

		cfg := mustGet(t, nil)

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

		mustChdir(t, projDir)
		err := exec.Command("git", "init").Run()
		if err != nil {
			t.Fatalf("failed to init a pseudo project: %s", err)
		}

		fs := pflag.NewFlagSet("test", pflag.ExitOnError)
		fs.String("port", "", "")
		fs.StringToString("header", nil, "")
		fs.StringSlice("path", nil, "")
		// --port flag changes port number to '8080'. Also --header appends 'foo=bar' and 'hoge=fuga' to 'request.header'.
		_ = fs.Parse([]string{
			"--port", "8080",
			"--path", "yoko.touma",
			"--header", "foo=bar", "--header", "hoge=fuga",
		})

		cfg := mustGet(t, fs)

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
		_ = fs.Parse([]string{"--port", "8080"})

		cfg := mustGet(t, fs)

		checkValues(t, cfg)

		return cfg
	})

	assertWithGolden(t, "apply some proto files and paths", func(t *testing.T) *Config {
		_, _, cleanup := setupEnv(t)
		defer cleanup()

		fs := pflag.NewFlagSet("test", pflag.ExitOnError)
		fs.StringSlice("path", []string{"foo", "bar"}, "")
		fs.StringSlice("proto", []string{"hoge", "fuga"}, "")

		cfg := mustGet(t, fs)

		checkValues(t, cfg)

		return cfg
	})
}

func TestEdit(t *testing.T) {
	cases := map[string]struct {
		outsideGitRepo bool
		runEditorErr   error // If it isn't nil, runEditor returns it.
		expectedEditor string
		hasErr         bool
	}{
		"run with default editor": {},
		"run with $EDITOR":        {expectedEditor: "nvim"},
		"Edit returns an error because it is outside a Git repo": {outsideGitRepo: true, hasErr: true},
	}

	p, err := exec.LookPath("vim")
	if err != nil {
		t.Fatalf("TestEdit requires Vim")
	}
	expected := p // default value
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			t.Setenv("EDITOR", c.expectedEditor)

			if c.outsideGitRepo {
				wd, err := os.Getwd()
				if err != nil {
					t.Fatalf("failed to get the working dir: %s", err)
				}
				dir := os.TempDir()
				mustChdir(t, dir)
				defer mustChdir(t, wd)
				defer os.Remove(dir)
			}

			var called bool
			runEditor = func(editor string, cfgPath string) error {
				expected := expected
				called = true
				if c.expectedEditor != "" {
					expected = c.expectedEditor
				}
				if expected != editor {
					t.Errorf("runEditor must be called with the expected editor (expected = %s, actual = %s)", expected, editor)
				}
				return c.runEditorErr
			}

			err := Edit()
			if c.hasErr {
				if err == nil {
					t.Error("Edit must return an error, but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Edit must not return an error, but got '%s'", err)
			}

			if !called {
				t.Error("runEditor must be called")
			}
		})
	}
}

func TestEditGlobal(t *testing.T) {
	cases := map[string]struct {
		runEditorErr   error // If it isn't nil, runEditor returns it.
		expectedEditor string
		hasErr         bool
	}{
		"run with default editor": {},
		"run with $EDITOR":        {expectedEditor: "nvim"},
	}

	p, err := exec.LookPath("vim")
	if err != nil {
		t.Fatalf("TestEdit requires Vim")
	}
	expected := p // default value
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			t.Setenv("EDITOR", c.expectedEditor)

			var called bool
			runEditor = func(editor string, cfgPath string) error {
				expected := expected
				called = true
				if c.expectedEditor != "" {
					expected = c.expectedEditor
				}
				if expected != editor {
					t.Errorf("runEditor must be called with the expected editor (expected = %s, actual = %s)", expected, editor)
				}
				return c.runEditorErr
			}

			err := EditGlobal()
			if c.hasErr {
				if err == nil {
					t.Error("Edit must return an error, but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Edit must not return an error, but got '%s'", err)
			}

			if !called {
				t.Error("runEditor must be called")
			}
		})
	}
}

func getWorkDir(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get the working dir: %s", err)
	}
	return cwd
}

func copyFile(t *testing.T, to, from string) {
	t.Helper()

	tf, err := os.Create(to)
	if err != nil {
		t.Fatalf("failed to create a config file: %s", err)
	}
	defer tf.Close()
	ff, err := os.Open(from)
	if err != nil {
		t.Fatalf("failed to open a prepared config file: %s", err)
	}
	defer ff.Close()
	if _, err := io.Copy(tf, ff); err != nil {
		t.Fatalf("io.Copy must not return an error, but got '%s'", err)
	}
}

func mkdir(t *testing.T, dir string) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("failed to create dirs: %s", err)
	}
}

func mustGet(t *testing.T, fs *pflag.FlagSet) *Config {
	cfg, err := Get(fs)
	if err != nil {
		t.Fatalf("Get must not return any errors, but got '%s'", err)
	}
	return cfg
}

func mustChdir(t *testing.T, dir string) {
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir must not return an error, but got '%s'", err)
	}
}
