package meta

import (
	"os"
	"path/filepath"

	semver "github.com/ktr0731/go-semver"
	updater "github.com/ktr0731/go-updater"
	homedir "github.com/mitchellh/go-homedir"
	toml "github.com/pelletier/go-toml"
)

const AppName = "evans"

var (
	Version = semver.MustParse("0.2.5")

	m               Meta
	defaultFileName = "meta.toml"
)

type MeansType updater.MeansType

const MeansTypeUndefined MeansType = ""

type Meta struct {
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

	if err := toml.NewDecoder(f).Decode(&m); err != nil {
		panic(err)
	}
}

func Get() *Meta {
	return &m
}

func Clear() error {
	m.UpdateAvailable = false
	m.LatestVersion = ""
	return save()
}

func SetUpdateInfo(latest *semver.Version) error {
	m.UpdateAvailable = true
	m.LatestVersion = latest.String()
	return save()
}

func SetInstalledBy(mt MeansType) error {
	m.InstalledBy = mt
	return save()
}

func resolvePath() (string, error) {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		home, err := homedir.Dir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".cache", AppName, defaultFileName), nil
	}
	return filepath.Join(base, AppName, defaultFileName), nil
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
	return toml.NewEncoder(f).Encode(Meta{
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
	return toml.NewEncoder(f).Encode(m)
}
