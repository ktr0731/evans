package meta

import (
	"os"
	"path/filepath"

	semver "github.com/ktr0731/go-semver"
	homedir "github.com/minio/go-homedir"
	toml "github.com/pelletier/go-toml"
)

const Name = "evans"

var (
	Version = semver.MustParse("0.2.8")

	m               Meta
	defaultFileName = "meta.toml"
)

type Meta struct {
	UpdateAvailable bool   `default:"false" toml:"updateAvailable"`
	LatestVersion   string `default:"" toml:"latestVersion"`
}

func init() {
	setup()
}

func setup() {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}
		base = filepath.Join(home, ".cache", Name)
	} else {
		base = filepath.Join(base, Name)
	}

	fname := filepath.Join(base, defaultFileName)

	if _, err := os.Stat(base); os.IsNotExist(err) {
		if err := initConfigFile(fname); err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	f, err := os.Open(fname)
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

func initConfigFile(fname string) error {
	if err := os.MkdirAll(filepath.Dir(fname), 0755); err != nil {
		return err
	}
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(Meta{
		UpdateAvailable: false,
		LatestVersion:   "",
	})
}
