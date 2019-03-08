package cache

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/go-updater/github"
	homedir "github.com/mitchellh/go-homedir"
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

func Test_resolvePath(t *testing.T) {
	t.Run("$XDG_CACHE_HOME is empty", func(t *testing.T) {
		cleanup := setEnv("XDG_CACHE_HOME", "")
		defer cleanup()

		home, err := homedir.Dir()
		require.NoError(t, err, "homedir.Dir must return the home dir")
		var expected string
		if runtime.GOOS == "windows" {
			expected = filepath.Join(filepath.FromSlash(os.Getenv("LOCALAPPDATA")), "cache", meta.AppName, defaultFileName)
		} else {
			expected = filepath.Join(home, ".cache", meta.AppName, defaultFileName)
		}

		actual := resolvePath()
		assert.Equal(t, expected, actual, "resolvePath must return $HOME/.cache as the cache dir")
	})

	t.Run("$XDG_CACHE_HOME is not empty", func(t *testing.T) {
		cleanup := setEnv("XDG_CACHE_HOME", testDir)
		defer cleanup()

		expected := filepath.Join(testDir, meta.AppName, defaultFileName)

		actual := resolvePath()
		assert.Equalf(t, expected, actual, "resolvePath must return %s as the cache dir", testDir)
	})
}

func TestCache(t *testing.T) {
	cleanup := setEnv("XDG_CACHE_HOME", testDir)
	defer cleanup()

	t.Run("cache does not exist", func(t *testing.T) {
		defer func() {
			os.RemoveAll(testDir)
			cachedCache = nil // See cachedCache comments for its behavior.
		}()
		assert.NotNil(t, Get(), "after setup called, cache file is written in $XDG_CACHE_HOME/testDir but returned cache was nil")
	})

	t.Run("cache exists", func(t *testing.T) {
		defer func() {
			os.RemoveAll(testDir)
			cachedCache = nil
		}()

		cache := Get()

		mt := MeansType(github.MeansTypeGitHubRelease)
		setCache := func() *Cache {
			cache := Get()

			unsavedCache := cache.SetUpdateInfo(version.Must(version.NewSemver("1.0.0")))
			assert.Equal(t, "1.0.0", unsavedCache.UpdateInfo.LatestVersion)
			assert.True(t, unsavedCache.UpdateInfo.UpdateAvailable)

			unsavedCache = cache.SetInstalledBy(mt)
			assert.Equal(t, unsavedCache.InstalledBy, mt)

			return unsavedCache
		}

		// Call setCache without calling Save.
		setCache()

		// get new cache value
		newCache := Get()
		assert.Equal(t, cache, newCache, "newCache must not modified")

		// Call setCache with calling Save.
		err := setCache().Save()
		require.NoError(t, err)

		newCache = Get()

		assert.Equal(t, "1.0.0", newCache.UpdateInfo.LatestVersion)
		assert.True(t, newCache.UpdateInfo.UpdateAvailable)
		assert.Equal(t, newCache.InstalledBy, mt)

		err = newCache.ClearUpdateInfo()
		require.NoError(t, err)
		newCache = Get()
		assert.Empty(t, newCache.UpdateInfo, "UpdateInfo must be cleared by ClearUpdateInfo, but got non-empty values")
	})
}
