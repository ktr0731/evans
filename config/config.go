package config

import (
	"github.com/kelseyhightower/envconfig"
	configure "github.com/ktr0731/go-configure"
	"github.com/mitchellh/mapstructure"
)

var mConfig *configure.Configure

type Server struct {
	Host string `default:"127.0.0.1" toml:"host"`
	Port string `default:"50051" toml:"port"`
}

type REPL struct {
	Server       *Server `toml:"-"`
	PromptFormat string  `default:"{package}.{sevice}@{addr}:{port}" toml:"prompt"`
	Reader       string  `default:"stdin" toml:"reader"`
	Writer       string  `default:"stdout" toml:"writer"`
	ErrWriter    string  `default:"stderr" toml:"err_writer"`
}

type Env struct {
	Server *Server `toml:"-"`
}

type Meta struct {
	Path string `default:"~/.config/evans/config.toml" toml:"path"`
}

type Config struct {
	Meta   *Meta   `toml:"meta"`
	REPL   *REPL   `toml:"repl"`
	Env    *Env    `toml:"env"`
	Server *Server `toml:"server"`
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
