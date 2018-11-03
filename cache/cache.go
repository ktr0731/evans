package cache

import (
	"os"
	"path/filepath"

	"github.com/ktr0731/evans/meta"
	semver "github.com/ktr0731/go-semver"
	updater "github.com/ktr0731/go-updater"
	homedir "github.com/mitchellh/go-homedir"
	toml "github.com/pelletier/go-toml"
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

func init() {
	setup()
}

func setup() {
	p, err := resolvePath()
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(filepath.Dir(p)); os.IsNotExist(err) {
		if err := initCacheFile(p); err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

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

	if err := toml.NewDecoder(f).Decode(&c); err != nil {
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
	return save()
}

// SetUpdateInfo sets an updatable flag to true and
// the latest version info to passed version.
func SetUpdateInfo(latest *semver.Version) error {
	c.UpdateAvailable = true
	c.LatestVersion = latest.String()
	return save()
}

// SetInstalledBy sets means how Evans was installed.
func SetInstalledBy(mt MeansType) error {
	c.InstalledBy = mt
	return save()
}

func resolvePath() (string, error) {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		home, err := homedir.Dir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".cache", meta.AppName, defaultFileName), nil
	}
	return filepath.Join(base, meta.AppName, defaultFileName), nil
}

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

func save() error {
	p, err := resolvePath()
	if err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}
