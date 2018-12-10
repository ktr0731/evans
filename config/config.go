package config

import (
	"encoding/csv"
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

// type Header map[string]string

type Request struct {
	Header map[string]string `toml:"header"`
	Web    bool              `toml:"web"`
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

	viper.SetDefault("request.header", map[string]string{"grpc-client": "evans"})
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
		switch f.Value.Type() {
		case "stringToString":
			// There is pflag.StringToString which converts 'key=val' to a map structure.
			// However, currently, we don't use BindPFlag because it has some bugs.
			currentMap := viper.GetStringMapString(k)
			newMap := stringToStringToMap(f.Value.String())
			for k, v := range newMap {
				currentMap[k] = v
			}
			viper.Set(k, currentMap)
			continue
		case "stringSlice":
			// We want to append flag values to the config.
			// So, we don't use BindPFlag.
			currentSlice := viper.GetStringSlice(k)
			newSlice := stringSliceToSlice(f.Value.String())
			for _, v := range newSlice {
				currentSlice = append(currentSlice, v)
			}
			viper.Set(k, currentSlice)
			continue
		}
		viper.BindPFlag(k, f)
	}
}

// stringToStringToMap convets (pflag.stringToStringValue).String() to a map.
// If some errors occur, stringToStringToMap returns an empty map.
func stringToStringToMap(val string) map[string]string {
	val = strings.Trim(val, "[]")
	if len(val) == 0 {
		return nil
	}
	r := csv.NewReader(strings.NewReader(val))
	ss, err := r.Read()
	if err != nil {
		return nil
	}
	out := make(map[string]string, len(ss))
	for _, pair := range ss {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return nil
		}
		out[kv[0]] = kv[1]
	}
	return out
}

func stringSliceToSlice(val string) []string {
	val = val[1 : len(val)-1]
	if len(val) == 0 {
		return nil
	}
	cr := csv.NewReader(strings.NewReader(val))
	records, err := cr.Read()
	if err != nil {
		return nil
	}
	return records
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

func initConfig(fs *pflag.FlagSet) (cfg *Config, err error) {
	defer func() {
		if err == nil {
			setupConfig(cfg)
		}
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
		return nil, errors.Wrap(err, "failed to unmarshal the config which is applied flag values")
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
	p, found := getLocalConfigPath()
	if !found {
		root, found := lookupProjectRootPath()
		if !found {
			return errors.New("--edit must be call inside a Git project")
		}
		p = filepath.Join(root, localConfigName)
		if err := viper.WriteConfigAs(p); err != nil {
			return errors.Wrapf(err, "failed to write the current config to %s", p)
		}
	}
	editor := getEditor()
	if editor == "" {
		return errors.New("--edit requires one of $EDITOR value or Vim")
	}
	cmd := exec.Command(editor, p)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to execute %s", editor)
	}
	return nil
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

func getEditor() string {
	if env := os.Getenv("EDITOR"); env != "" {
		return env
	}
	p, err := exec.LookPath("vim")
	if err != nil {
		return ""
	}
	return p
}
