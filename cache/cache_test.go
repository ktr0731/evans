package cache

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestMain(m *testing.M) {
	CachedCache = nil

	// Create a temp dir to reserve dir name. But remove it, and create again later.
	dir := os.TempDir()

	oldCacheDir := os.Getenv("XDG_CACHE_HOME")
	os.Setenv("XDG_CACHE_HOME", dir)

	code := m.Run()

	os.RemoveAll(filepath.Dir(resolvePath()))
	os.Setenv("XDG_CACHE_HOME", oldCacheDir)
	CachedCache = nil

	os.Exit(code)
}

func TestCache(t *testing.T) {
	t.Run("Get creates a new one", func(t *testing.T) {
		CachedCache = nil
		_, err := Get()
		if err != nil {
			t.Fatalf("Get must not return an error, but got '%s'", err)
		}
	})

	t.Run("the file is empty", func(t *testing.T) {
		CachedCache = nil
		oldTOMLDecoder := tomlDecodeReader
		defer func() {
			tomlDecodeReader = oldTOMLDecoder
		}()
		// Do nothing.
		tomlDecodeReader = func(r io.Reader, v interface{}) (toml.MetaData, error) {
			return toml.MetaData{}, nil
		}
		_, err := Get()
		if err != nil {
			t.Fatalf("Get must not return an error, but got '%s'", err)
		}
	})

	t.Run("Save", func(t *testing.T) {
		CachedCache = nil
		c, err := Get()
		if err != nil {
			t.Fatalf("Get must not return an error, but got '%s'", err)
		}
		c.Version = "5.0.0"
		if err := c.Save(); err != nil {
			t.Fatalf("must not return an error, but got '%s'", err)
		}
	})
}
