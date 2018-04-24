package config

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
	configure "github.com/ktr0731/go-configure"
	"github.com/ktr0731/mapstruct"
	"github.com/mitchellh/mapstructure"
)

var (
	localConfigName = ".evans.toml"
)

var mConfig *configure.Configure

type Server struct {
	Host string `default:"127.0.0.1" toml:"host"`
	Port string `default:"50051" toml:"port"`
}

type Header struct {
	Key string `toml:"key"`
	Val string `toml:"val"`
}

type Request struct {
	Header []Header `toml:"header"`
}

type REPL struct {
	Server       *Server `toml:"-"`
	PromptFormat string  `default:"{package}.{sevice}@{addr}:{port}" toml:"promptFormat"`
	Reader       string  `default:"stdin" toml:"reader"`
	Writer       string  `default:"stdout" toml:"writer"`
	ErrWriter    string  `default:"stderr" toml:"errWriter"`

	ColoredOutput bool `default:"true" toml:"coloredOutput"`

	ShowSplashText bool   `default:"true" toml:"showSplashText"`
	SplashTextPath string `default:"" toml:"splashTextPath"`
}

type Env struct {
	Server *Server `toml:"-"`
}

type Input struct {
	PromptFormat string `default:"{ancestor}{name} ({type}) => " toml:"promptFormat"`
}

type Meta struct {
	Path        string `default:"~/.config/evans/config.toml" toml:"path"`
	AutoUpdate  bool   `default:"false" toml:"autoUpdate"`
	UpdateLevel string `default:"patch" toml:"updateLevel"`
}

type Config struct {
	Default *Default `toml:"default"`
	Meta    *Meta    `toml:"meta"`
	REPL    *REPL    `toml:"repl"`
	Env     *Env     `toml:"env"`
	Server  *Server  `toml:"server"`
	Log     *Log     `toml:"log"`
	Request *Request `toml:"request"`
	Input   *Input   `toml:"input"`
}

type Default struct {
	ProtoPath []string `toml:"protoPath" default:""`
	ProtoFile []string `toml:"protoFile" default:""`
	Package   string   `toml:"package" default:""`
	Service   string   `toml:"service" default:""`
}

type Log struct {
	Prefix string `default:"[evans] " toml:"prefix"`
}

type localConfig struct {
	Default *Default `toml:"default"`
}

func init() {
	conf := Config{
		Request: &Request{
			Header: []Header{
				{"user-agent", "evans"},
			},
		},
		// to show items in initial config file, set an empty value
		Default: &Default{
			ProtoPath: []string{""},
			ProtoFile: []string{""},
		},
	}
	var err error
	if err := envconfig.Process("evans", &conf); err != nil {
		panic(err)
	}

	mConfig, err = configure.NewConfigure(conf.Meta.Path, conf, nil)
	if err != nil {
		panic(err)
	}
}

func setupConfig(c *Config) {
	c.REPL.Server = c.Server
	c.Env.Server = c.Server
}

func Get() *Config {
	var global Config
	err := mapstructure.Decode(mConfig.Get(), &global)
	if err != nil {
		panic(err)
	}

	local, err := getLocalConfig()
	if err != nil {
		panic(err)
	}

	// if local config missing, return global config
	if local == nil {
		setupConfig(&global)
		return &global
	}

	ic, err := mapstruct.Map(&global, local)
	if err != nil {
		panic(err)
	}

	c := ic.(*Config)
	setupConfig(c)

	return c
}

func Edit() error {
	return mConfig.Edit()
}

func getLocalConfig() (*Config, error) {
	var f io.ReadCloser
	if _, err := os.Stat(localConfigName); err != nil {
		if os.IsNotExist(err) {
			f, err = lookupProjectRoot()
			// local file not found
			if f == nil || err != nil {
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
	var conf Config
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
