package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ktr0731/evans/logger"
	configure "github.com/ktr0731/go-configure"
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
	Host       string `toml:"host"`
	Port       string `toml:"port"`
	Reflection bool   `toml:"reflection"`
}

type Header map[string]string

type Request struct {
	Header Header `toml:"header"`
	Web    bool   `toml:"web"`
}

type REPL struct {
	// TODO: remove this
	Server       *Server `toml:"server"`
	PromptFormat string  `toml:"promptFormat"`

	ColoredOutput bool `toml:"coloredOutput"`

	ShowSplashText bool   `toml:"showSplashText"`
	SplashTextPath string `toml:"splashTextPath"`
}

type Meta struct {
	AutoUpdate  bool   `toml:"autoUpdate"`
	UpdateLevel string `toml:"updateLevel"`
}

type Config struct {
	Default *Default `toml:"default"`
	Meta    *Meta    `toml:"meta"`
	REPL    *REPL    `toml:"repl"`
	Server  *Server  `toml:"server"`
	Log     *Log     `toml:"log"`
	Request *Request `toml:"request"`
}

type Default struct {
	ProtoPath []string `toml:"protoPath"`
	ProtoFile []string `toml:"protoFile"`
	Package   string   `toml:"package"`
	Service   string   `toml:"service"`
}

type Log struct {
	Prefix string `toml:"prefix"`
}

func Get(fs *pflag.FlagSet) (*Config, error) {
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

	viper.SetDefault("request.header", Header{"grpc-client": "evans"})
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
			logger.Printf("flag is not found: %s-%s", k, v)
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
	setupConfig(&cfg)
	return &cfg, nil
}

// TODO: sync.Once
func initConfig(fs *pflag.FlagSet) (cfg *Config, err error) {
	defer func() {
		setupConfig(cfg)
	}()

	initDefaultValues()

	cfgDir := filepath.Join(xdgbasedir.ConfigHome(), "evans")

	// Global config paths
	viper.SetConfigType("toml")
	viper.SetConfigName("config")
	viper.AddConfigPath(cfgDir)

	logger.Printf("load global config from %s", cfgDir)
	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			path := filepath.Join(cfgDir, globalConfigName)
			logger.Printf("global config is not found, create a new one: %s", path)
			if err := os.MkdirAll(cfgDir, 0755); err != nil {
				return nil, errors.Wrap(err, "failed to create config dirs")
			}
			if err := viper.WriteConfigAs(path); err != nil {
				return nil, errors.Wrap(err, "failed to write a default config")
			}
			cfg, err = defaultConfig()
			return
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
		logger.Println("local config is not found")
		cfg = &globalCfg
		return
	}

	logger.Printf("load local config from %s", p)
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
		logger.Println("flagset is not found")
		cfg = &mergedCfg
		return
	}

	logger.Println("bind flagset to the loaded config")
	bindFlags(fs)
	var finalCfg Config
	if err := viper.Unmarshal(&finalCfg); err != nil {
		return nil, err
	}
	cfg = &finalCfg
	return
}

func setupConfig(c *Config) {
	// To show protofile and protopath field, set slice which has empty string
	// if these are nil. (please see default values.)
	// Conversely, trim the empty string element when config loading.
	if (c.Default.ProtoFile == nil) || (len(c.Default.ProtoFile) == 1 && c.Default.ProtoFile[0] == "") {
		c.Default.ProtoFile = []string{}
	}
	if (c.Default.ProtoPath == nil) || (len(c.Default.ProtoPath) == 1 && c.Default.ProtoPath[0] == "") {
		c.Default.ProtoPath = []string{}
	}
	c.REPL.Server = c.Server
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

func lookupProjectRootPath() (string, bool) {
	b, err := exec.Command("git", "rev-parse", "--show-cdup").Output()
	if err != nil {
		return "", false
	}
	p := strings.TrimSpace(string(b))
	return p, p != ""
}
