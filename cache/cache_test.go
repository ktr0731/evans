package cache

import (
	"io"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Create a temp dir to reserve dir name. But remove it, and create again later.
	dir := os.TempDir()

	oldCacheDir := os.Getenv("XDG_CACHE_HOME")
	os.Setenv("XDG_CACHE_HOME", dir)

	code := m.Run()

	os.Setenv("XDG_CACHE_HOME", oldCacheDir)
	os.Exit(code)
}

func TestCache(t *testing.T) {
	t.Run("Get creates a new one", func(t *testing.T) {
		_, err := Get()
		if err != nil {
			t.Fatalf("Get must not return an error, but got '%s'", err)
		}
	})

	t.Run("the file is empty", func(t *testing.T) {
		oldDecodeTOML := decodeTOML
		defer func() {
			decodeTOML = oldDecodeTOML
		}()
		// Do nothing.
		decodeTOML = func(r io.Reader, v interface{}) error {
			return nil
		}
		_, err := Get()
		if err != nil {
			t.Fatalf("Get must not return an error, but got '%s'", err)
		}
	})

	t.Run("Save", func(t *testing.T) {
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
