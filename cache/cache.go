package cache

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ktr0731/evans/meta"
	semver "github.com/ktr0731/go-semver"
	updater "github.com/ktr0731/go-updater"
	xdgbasedir "github.com/zchee/go-xdgbasedir"
)

var (
	c               Cache
	defaultFileName = "cache.toml"
)

type MeansType updater.MeansType

const MeansTypeUndefined MeansType = ""

// Cache represents cached items.
type Cache struct {
	UpdateAvailable bool      `default:"false" toml:"updateAvailable"`
	LatestVersion   string    `default:"" toml:"latestVersion"`
	InstalledBy     MeansType `default:"" toml:"installedBy"`
}

// Save writes the receiver to the cache file.
// It returns an *os.PathError if it can't create a new cache file.
// Also it returns an error if it failed to encode *Cache with TOML format.
func (c *Cache) Save() error {
	p := resolvePath()

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}

func init() {
	setup()
}

func setup() {
	c = Cache{}

	p := resolvePath()

	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := initCacheFile(p); err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	f, err := os.Open(p)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := toml.DecodeReader(f, &c); err != nil {
		panic(err)
	}
}

// Get returns loaded cache contents.
// Returned *Cache is NOT goroutine safe.
func Get() *Cache {
	return &c
}

// Clear clears contents of the cache file.
func Clear() error {
	c.UpdateAvailable = false
	c.LatestVersion = ""
	return c.Save()
}

// SetUpdateInfo sets an updatable flag to true and
// the latest version info to passed version.
func SetUpdateInfo(latest *semver.Version) *Cache {
	c.UpdateAvailable = true
	c.LatestVersion = latest.String()
	c2 := c
	return &c2
}

// SetInstalledBy sets means how Evans was installed.
func SetInstalledBy(mt MeansType) *Cache {
	c.InstalledBy = mt
	c2 := c
	return &c2
}

func resolvePath() string {
	return filepath.Join(xdgbasedir.CacheHome(), meta.AppName, defaultFileName)
}

// initCacheFile creates or overwrites a new cache file with default values.
// If directories of the file are not found, initCacheFile also creates it.
func initCacheFile(p string) error {
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(Cache{
		UpdateAvailable: false,
		LatestVersion:   "",
	})
}
