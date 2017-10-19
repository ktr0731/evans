package config

import (
	"github.com/kelseyhightower/envconfig"
	configure "github.com/ktr0731/go-configure"
	"github.com/mitchellh/mapstructure"
)

var mConfig *configure.Configure

type Core struct {
	Port string `default:"50051" toml:"port"`
}

type Meta struct {
	Path         string `default:"~/.config/evans/config.toml" toml:"path"`
	PromptFormat string `default:"{package}.{sevice}@{addr}:{port}" toml:"prompt"`
}

type Config struct {
	Meta *Meta `toml:"meta"`
	Core *Core `toml:"core"`
}

func init() {
	var conf Config
	var err error
	if err := envconfig.Process("evans", &conf); err != nil {
		panic(err)
	}
	mConfig, err = configure.NewConfigure(conf.Meta.Path, conf, nil)
	if err != nil {
		panic(err)
	}
}

func Get() *Config {
	var config Config
	err := mapstructure.Decode(mConfig.Get(), &config)
	if err != nil {
		panic(err)
	}
	return &config
}

func Edit() error {
	return mConfig.Edit()
}
