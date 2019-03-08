package cache

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/go-version"
	"github.com/ktr0731/evans/meta"
	updater "github.com/ktr0731/go-updater"
	xdgbasedir "github.com/zchee/go-xdgbasedir"
)

const defaultFileName = "cache.toml"

type MeansType updater.MeansType

const MeansTypeUndefined MeansType = ""

type UpdateInfo struct {
	UpdateAvailable bool   `default:"false" toml:"updateAvailable"`
	LatestVersion   string `default:"" toml:"latestVersion"`
}

// Cache represents cached items.
type Cache struct {
	Version        string     `toml:"version"`
	UpdateInfo     UpdateInfo `toml:"updateInfo"`
	InstalledBy    MeansType  `default:"" toml:"installedBy"`
	CommandHistory []string   `default:"" toml:"commandHistory"`
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
	cachedCache = c
	return toml.NewEncoder(f).Encode(c)
}

// cachedCache holds a loaded cache instantiated by Get.
// If cachedCache isn't nil, Get returns it straightforwardly.
var cachedCache *Cache

// Get returns loaded cache contents. To reduce duplicatd function calls,
// Get caches the result of Get. See cachedCache comments for more implementation details.
func Get() *Cache {
	if cachedCache != nil {
		return cachedCache
	}

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

	var c Cache
	if _, err := toml.DecodeReader(f, &c); err != nil {
		panic(err)
	}

	// If c.Version is empty or not equal to the latest version,
	// it is regarded as an old version.
	// In such case, we discard the loaded cache.
	if c.Version == "" || c.Version != meta.Version.String() {
		if err := initCacheFile(p); err != nil {
			panic(err)
		}
		if _, err := f.Seek(0, 0); err != nil {
			panic(err)
		}
		if _, err := toml.DecodeReader(f, &c); err != nil {
			panic(err)
		}
	}

	cachedCache = &c

	return &c
}

// ClearUpdateInfo clears c.UpdateInfo.
// ClearUpdateInfo also saves cleared cache to the file.
func (c *Cache) ClearUpdateInfo() error {
	c.UpdateInfo = UpdateInfo{
		UpdateAvailable: false,
		LatestVersion:   "",
	}
	return c.Save()
}

// SetUpdateInfo sets an updatable flag to true and
// the latest version info to passed version.
func (c *Cache) SetUpdateInfo(latest *version.Version) *Cache {
	c.UpdateInfo = UpdateInfo{
		UpdateAvailable: true,
		LatestVersion:   latest.String(),
	}
	return c
}

// SetInstalledBy sets means how Evans was installed.
func (c *Cache) SetInstalledBy(mt MeansType) *Cache {
	c.InstalledBy = mt
	return c
}

func (c *Cache) SetCommandHistory(h []string) *Cache {
	c.CommandHistory = h
	return c
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
		Version: meta.Version.String(),
	})
}
