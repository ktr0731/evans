package cache

import (
	"os"
	"testing"

	semver "github.com/ktr0731/go-semver"
	"github.com/ktr0731/go-updater/github"
	toml "github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

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
		assert.NotNil(t, Get(), "after setup called, cache file is written in $XDG_CACHE_HOME/testDir but returned cache was nil")
	})

	t.Run("config exist", func(t *testing.T) {
		setup()
		defer func() {
			os.RemoveAll(testDir)
		}()

		cache := Get()
		err := SetUpdateInfo(semver.MustParse("1.0.0"))
		require.NoError(t, err)
		mt := MeansType(github.MeansTypeGitHubRelease)
		err = SetInstalledBy(mt)
		require.NoError(t, err)

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
		assert.Equal(t, newCache.InstalledBy, MeansType(mt))
	})
}
