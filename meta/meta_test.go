package meta

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testDir = "tmp"

func setEnv(k, v string) func() {
	old := os.Getenv(k)
	os.Setenv(k, v)
	return func() {
		os.Setenv(k, old)
	}
}

func TestMeta(t *testing.T) {
	cleanup := setEnv("XDG_CACHE_HOME", testDir)
	defer cleanup()

	t.Run("config not exist", func(t *testing.T) {
		setup()
		defer func() {
			os.RemoveAll(testDir)
		}()
	})

	t.Run("config exist", func(t *testing.T) {
		setup()
		setup()
		defer func() {
			os.RemoveAll(testDir)
		}()
	})

	assert.NotNil(t, Get())
}
