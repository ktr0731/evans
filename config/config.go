package config

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	configure "github.com/ktr0731/go-configure"
	"github.com/ktr0731/mapstruct"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	xdgbasedir "github.com/zchee/go-xdgbasedir"
)

var (
	localConfigName  = ".evans.toml"
	globalConfigName = "config.toml"
)

var mConfig *configure.Configure

type Server struct {
	Host       string `default:"127.0.0.1" toml:"host"`
	Port       string `default:"50051" toml:"port"`
	Reflection bool   `default:"false" toml:"reflection"`
}

type Header struct {
	Key string `toml:"key"`
	Val string `toml:"val"`
}

type Header2 map[string]string

type Request struct {
	Header Header2 `toml:"header"`
	Web    bool    `toml:"web"`
}

type REPL struct {
	Server       *Server `toml:"-"`
	PromptFormat string  `toml:"promptFormat"`

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

func Get2(fs *pflag.FlagSet) (*Config, error) {
	return initConfig(fs)
}

func initDefaultValues() {
	viper.SetDefault("default.protoPath", []string{""})
	viper.SetDefault("default.protoFile", []string{""})
	viper.SetDefault("default.package", "")
	viper.SetDefault("default.service", "")

	viper.SetDefault("meta.autoUpdate", false)
	viper.SetDefault("meta.updateLevel", "patch")

	viper.SetDefault("repl.promptFormat", "{package}.{sevice}@{addr}:{port}")
	viper.SetDefault("repl.coloredOutput", true)
	viper.SetDefault("repl.showSplashText", true)
	viper.SetDefault("repl.splashTextPath", "")

	viper.SetDefault("server.host", "127.0.0.1")
	viper.SetDefault("server.port", "50051")
	viper.SetDefault("server.reflection", false)

	viper.SetDefault("log.prefix", "evans: ")

	viper.SetDefault("request.header", Header2{"grpc-client": "evans"})
	viper.SetDefault("request.web", false)
}

func bindFlags(fs *pflag.FlagSet) {
	kv := map[string]string{
		"default.protoPath": "path",
		"default.protoFile": "file",
		"default.package":   "package",
		"default.service":   "service",
		"server.host":       "host",
		"server.port":       "port",
		"server.reflection": "reflection",
		"request.header":    "header",
		"request.web":       "web",
	}
	for k, v := range kv {
		f := fs.Lookup(v)
		if f == nil {
			continue
		}
		viper.BindPFlag(k, f)
	}
}

func defaultConfig() (*Config, error) {
	viper.Reset()
	initDefaultValues()
	var cfg Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal default config")
	}
	return &cfg, nil
}

// TODO: sync.Once
func initConfig(fs *pflag.FlagSet) (*Config, error) {
	initDefaultValues()

	cfgDir := filepath.Join(xdgbasedir.ConfigHome(), "evans")

	// Global config paths
	viper.SetConfigType("toml")
	viper.SetConfigName("config")
	viper.AddConfigPath(cfgDir)
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := os.MkdirAll(cfgDir, 0755); err != nil {
				return nil, errors.Wrap(err, "failed to create config dirs")
			}
			if err := viper.WriteConfigAs(filepath.Join(cfgDir, globalConfigName)); err != nil {
				return nil, errors.Wrap(err, "failed to write a default config")
			}
			return defaultConfig()
		} else {
			return nil, err
		}
	}
	var globalCfg Config
	if err := viper.Unmarshal(&globalCfg); err != nil {
		return nil, err
	}

	p, found := getLocalConfigPath()
	if !found {
		return &globalCfg, nil
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open a local config file")
	}
	defer f.Close()
	if err := viper.MergeConfig(f); err != nil {
		return nil, errors.Wrap(err, "failed to merge a local config to the global config")
	}

	var mergedCfg Config
	if err := viper.Unmarshal(&mergedCfg); err != nil {
		return nil, err
	}

	if fs == nil {
		return &mergedCfg, nil
	}

	bindFlags(fs)
	var finalCfg Config
	if err := viper.Unmarshal(&finalCfg); err != nil {
		return nil, err
	}
	return &finalCfg, nil
}

func SetupConfig(c *Config) {
	if len(c.Default.ProtoFile) == 1 && c.Default.ProtoFile[0] == "" {
		c.Default.ProtoFile = []string{}
	}
	if len(c.Default.ProtoPath) == 1 && c.Default.ProtoPath[0] == "" {
		c.Default.ProtoPath = []string{}
	}
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
		SetupConfig(&global)
		return &global
	}

	ic, err := mapstruct.Map(&global, local)
	if err != nil {
		panic(err)
	}

	c := ic.(*Config)
	SetupConfig(c)

	return c
}

func Edit() error {
	return mConfig.Edit()
}

func getLocalConfigPath() (string, bool) {
	if _, err := os.Stat(localConfigName); err != nil {
		if os.IsNotExist(err) {
			root, found := lookupProjectRootPath()
			if !found {
				return "", false
			}
			p := filepath.Join(root, localConfigName)
			if _, err := os.Stat(p); os.IsNotExist(err) {
				return "", false
			}
			return p, true
		}
		return "", false
	}
	p, err := filepath.Abs(localConfigName)
	return p, err == nil
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
	}

	f, err := os.Open(localConfigName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var conf Config
	_, err = toml.DecodeReader(f, &conf)
	return &conf, err
}

func lookupProjectRootPath() (string, bool) {
	b, err := exec.Command("git", "rev-parse", "--show-cdup").Output()
	if err != nil {
		return "", false
	}
	p := strings.TrimSpace(string(b))
	return p, p != ""
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
