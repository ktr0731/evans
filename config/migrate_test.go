package config

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

// Test_migration tests whether Get doesn't return errors.
func Test_migration(t *testing.T) {
	toFileName := func(ver string) string {
		return strings.Replace(ver, ".", "_", -1) + ".toml"
	}

	for oldVer := range migrationScripts {
		oldVer := oldVer
		t.Run(oldVer, func(t *testing.T) {
			oldCWD := getWorkDir(t)

			_, cfgDir, cleanup := setupEnv(t)
			defer cleanup()

			b, err := ioutil.ReadFile(filepath.Join(oldCWD, "testdata", toFileName(oldVer)))
			if err != nil {
				t.Fatalf("failed to read a config file, but got '%s'", err)
			}

			err = ioutil.WriteFile(filepath.Join(cfgDir, "config.toml"), b, 0600)
			if err != nil {
				t.Fatalf("failed to copy a config file to a temp config dir, but got '%s'", err)
			}

			_, err = Get(nil)
			if err != nil {
				t.Errorf("Get must not return errors, but got '%s'", err)
			}
		})
	}
}
