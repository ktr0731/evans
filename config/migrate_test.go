package config

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_migration(t *testing.T) {
	toFileName := func(ver string) string {
		return strings.Replace(ver, ".", "_", -1) + ".toml"
	}

	for oldVer := range migrationScripts {
		t.Run(oldVer, func(t *testing.T) {
			oldCWD := getWorkDir(t)

			_, cfgDir, cleanup := setupEnv(t)
			defer cleanup()

			b, err := ioutil.ReadFile(filepath.Join(oldCWD, "testdata", toFileName(oldVer)))
			require.NoError(t, err, "failed to read a config file")

			err = ioutil.WriteFile(filepath.Join(cfgDir, "config.toml"), b, 0644)
			require.NoError(t, err, "failed to copy a config file to a temp config dir")

			_, err = Get(nil)
			assert.NoError(t, err, "Get must not return errors")
		})
	}
}
