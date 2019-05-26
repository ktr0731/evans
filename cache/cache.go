package cache

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/go-updater"
	"github.com/pkg/errors"
	"github.com/zchee/go-xdgbasedir"
)

const defaultFileName = "cache.toml"

type MeansType updater.MeansType

const MeansTypeUndefined MeansType = ""

type UpdateInfo struct {
	LatestVersion string    `default:"" toml:"latestVersion"`
	InstalledBy   MeansType `default:"" toml:"installedBy"`
}

func (i UpdateInfo) UpdateAvailable() bool {
	return i.LatestVersion != ""
}

// Cache represents cached items.
type Cache struct {
	// TODO: いらない？
	Version        string     `toml:"version"`
	UpdateInfo     UpdateInfo `toml:"updateInfo"`
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
	cachedCacheMu.Lock()
	defer cachedCacheMu.Unlock()
	CachedCache = c
	return toml.NewEncoder(f).Encode(c)
}

var (
	cachedCacheMu sync.Mutex
	// CachedCache holds a loaded cache instantiated by Get.
	// If CachedCache isn't nil, Get returns it straightforwardly.
	CachedCache *Cache
)

var tomlDecodeReader = toml.DecodeReader

// Get returns loaded cache contents. To reduce duplicatd function calls,
// Get caches the result of Get. See CachedCache comments for more implementation details.
var Get = func() (*Cache, error) {
	if CachedCache != nil {
		return CachedCache, nil
	}

	p := resolvePath()

	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := initCacheFile(p); err != nil {
			return nil, errors.Wrap(err, "failed to create a new cache file")
		}
	} else if err != nil {
		return nil, errors.Wrapf(err, "failed to check the existency of cache file '%s'", p)
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open the cache file")
	}
	defer f.Close()

	var c Cache
	if _, err := tomlDecodeReader(f, &c); err != nil {
		return nil, errors.Wrap(err, "failed to decode loaded cache content")
	}

	// If c.Version is empty or not equal to the latest version,
	// it is regarded as an old version.
	// In such case, we clear the loaded cache.
	if c.Version == "" || c.Version != meta.Version.String() {
		if err := initCacheFile(p); err != nil {
			return nil, errors.Wrap(err, "failed to clear the cache file")
		}
		if _, err := f.Seek(0, 0); err != nil {
			return nil, errors.Wrap(err, "failed to move to the first")
		}
		if _, err := toml.DecodeReader(f, &c); err != nil {
			return nil, errors.Wrap(err, "failed to decode cache content")
		}
	}

	cachedCacheMu.Lock()
	defer cachedCacheMu.Unlock()
	CachedCache = &c

	return &c, nil
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
