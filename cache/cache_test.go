package cache

import (
	"os"
	"testing"

	toml "github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDir = "tmp"

func setEnv(k, v string) func() {
	old := os.Getenv(k)
	os.Setenv(k, v)
	return func() {
		os.Setenv(k, old)
	}
}

func TestCache(t *testing.T) {
	cleanup := setEnv("XDG_CACHE_HOME", testDir)
	defer cleanup()

	t.Run("config not exist", func(t *testing.T) {
		setup()
		defer func() {
			os.RemoveAll(testDir)
		}()
		assert.NotNil(t, Get())
	})

	t.Run("config exist", func(t *testing.T) {
		setup()
		defer func() {
			os.RemoveAll(testDir)
		}()

		cache := Get()
		cache.LatestVersion = "1.0.0"
		cache.UpdateAvailable = true

		p, err := resolvePath()
		require.NoError(t, err)
		f, err := os.Create(p)
		require.NoError(t, err)
		defer f.Close()

		err = toml.NewEncoder(f).Encode(*cache)
		require.NoError(t, err)

		// get new cached
		setup()

		newCache := Get()
		assert.Equal(t, "1.0.0", newCache.LatestVersion)
		assert.True(t, newCache.UpdateAvailable)
	})
}
