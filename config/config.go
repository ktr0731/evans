package config

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	configure "github.com/ktr0731/go-configure"
	"github.com/ktr0731/toml"
	"github.com/mitchellh/mapstructure"
)

const (
	localConfigName = ".evans.toml"
)

var mConfig *configure.Configure

type Server struct {
	Host string `default:"127.0.0.1" toml:"host"`
	Port string `default:"50051" toml:"port"`
}

type REPL struct {
	Server       *Server `toml:"-"`
	PromptFormat string  `default:"{package}.{sevice}@{addr}:{port}" toml:"promptFormat"`
	Reader       string  `default:"stdin" toml:"reader"`
	Writer       string  `default:"stdout" toml:"writer"`
	ErrWriter    string  `default:"stderr" toml:"errWriter"`

	SplashTextPath string `default:"" toml:"splashTextPath"`
}

type Env struct {
	Server            *Server `toml:"-"`
	AncestorDelimiter string  `default:":"`
	InputPromptFormat string  `default:"{name}{ancestor} ({type}) => " toml:"inputPromptFormat"`
}

type Meta struct {
	Path string `default:"~/.config/evans/config.toml" toml:"path"`
}

type Config struct {
	Default *Default `toml:"default"`
	Meta    *Meta    `toml:"meta"`
	REPL    *REPL    `toml:"repl"`
	Env     *Env     `toml:"env"`
	Server  *Server  `toml:"server"`
}

type Default struct {
	Package string `toml:"package"`
	Service string `toml:"service"`
}

type localConfig struct {
	Default *Default `toml:"default"`
}

func init() {
	var conf Config
	var err error
	if err := envconfig.Process("evans", &conf); err != nil {
		panic(err)
	}

	// TODO: use more better method
	conf.REPL.Server = conf.Server
	conf.Env.Server = conf.Server

	local, err := getLocalConfig()
	if err != nil {
		panic(err)
	}

	applyLocalConfig(&conf, local)

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

	// TODO: use more better method
	config.REPL.Server = config.Server
	config.Env.Server = config.Server

	local, err := getLocalConfig()
	if err != nil {
		panic(err)
	}
	applyLocalConfig(&config, local)

	return &config
}

func Edit() error {
	return mConfig.Edit()
}

func getLocalConfig() (*localConfig, error) {
	var f io.ReadCloser
	if _, err := os.Stat(localConfigName); err != nil {
		if os.IsNotExist(err) {
			f, err = lookupProjectRoot()
			if err != nil {
				return nil, err
			}
			// local file not found
			if f == nil {
				return nil, nil
			}
			defer f.Close()
		}
		return nil, err
	} else {
		f, err = os.Open(localConfigName)
		if err != nil {
			return nil, err
		}
		defer f.Close()
	}
	var conf localConfig
	_, err := toml.DecodeReader(f, &conf)
	return &conf, err
}

func lookupProjectRoot() (io.ReadCloser, error) {
	outBuf, errBuf := new(bytes.Buffer), new(bytes.Buffer)
	cmd := exec.Command("git", "rev-parse", "--show-cdup")
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	if errBuf.Len() != 0 {
		return nil, errors.New(errBuf.String())
	}
	p := filepath.Join(outBuf.String(), localConfigName)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil, nil
	}
	return os.Open(p)
}

func applyLocalConfig(global *Config, local *localConfig) {
	if local == nil {
		return
	}
	if local.Default.Package != "" {
		global.Default.Package = local.Default.Package
	}
	if local.Default.Service != "" {
		global.Default.Service = local.Default.Service
	}
}
