package config

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
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
	if err != nil {
		t.Fatalf("failed to create a temp dir to setup testing environment: %s", err)
	}

	mustChdir(t, dir)
	oldEnv := os.Getenv("XDG_CONFIG_HOME")
	cfgDir := filepath.Join(dir, "config")
	os.Setenv("XDG_CONFIG_HOME", cfgDir)
	evansCfgDir := filepath.Join(cfgDir, "evans")
	mkdir(t, evansCfgDir)

	return dir, evansCfgDir, func() {
		mustChdir(t, cwd)
		os.Setenv("XDG_CONFIG_HOME", oldEnv)
		os.RemoveAll(dir)
		viper.Reset()
	}
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
			oldEnv := os.Getenv("EDITOR")
			os.Setenv("EDITOR", c.expectedEditor)
			defer os.Setenv("EDITOR", oldEnv)

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
			oldEnv := os.Getenv("EDITOR")
			os.Setenv("EDITOR", c.expectedEditor)
			defer os.Setenv("EDITOR", oldEnv)

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

func mustChdir(t *testing.T, dir string) {
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir must not return an error, but got '%s'", err)
	}
}

func mkdir(t *testing.T, dir string) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("failed to create dirs: %s", err)
	}
}
