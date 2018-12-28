package config

import (
	"encoding/csv"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ktr0731/evans/logger"
	"github.com/ktr0731/evans/meta"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	xdgbasedir "github.com/zchee/go-xdgbasedir"
)

var (
	localConfigName  = ".evans.toml"
	globalConfigName = "config.toml"
)

type Server struct {
	Host       string `toml:"host"`
	Port       string `toml:"port"`
	Reflection bool   `toml:"reflection"`
	TLS        bool   `toml:"tls"`
}

type Header map[string][]string

type Request struct {
	Header Header `toml:"header"`
	Web    bool   `toml:"web"`
}

type REPL struct {
	PromptFormat      string `toml:"promptFormat"`
	InputPromptFormat string `toml:"inputPromptFormat"`

	ColoredOutput bool `toml:"coloredOutput"`

	ShowSplashText bool   `toml:"showSplashText"`
	SplashTextPath string `toml:"splashTextPath"`
}

type Meta struct {
	ConfigVersion string `toml:"configVersion"`
	AutoUpdate    bool   `toml:"autoUpdate"`
	UpdateLevel   string `toml:"updateLevel"`
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

func newDefaultViper() *viper.Viper {
	v := viper.New()
	v.SetDefault("default.protoPath", []string{""})
	v.SetDefault("default.protoFile", []string{""})
	v.SetDefault("default.package", "")
	v.SetDefault("default.service", "")

	// We set the default version to v0.6.10
	// because the structure of Config is changed at v0.6.11.
	v.SetDefault("meta.configVersion", "0.6.10")
	v.SetDefault("meta.autoUpdate", false)
	v.SetDefault("meta.updateLevel", "patch")

	v.SetDefault("repl.promptFormat", "{package}.{sevice}@{addr}:{port}")
	v.SetDefault("repl.inputPromptFormat", "{ancestor}{name} ({type}) => ")
	v.SetDefault("repl.coloredOutput", true)
	v.SetDefault("repl.showSplashText", true)
	v.SetDefault("repl.splashTextPath", "")

	v.SetDefault("server.host", "127.0.0.1")
	v.SetDefault("server.port", "50051")
	v.SetDefault("server.reflection", false)
	v.SetDefault("server.tls", false)

	v.SetDefault("log.prefix", "evans: ")

	v.SetDefault("request.header", Header{"grpc-client": []string{"evans"}})
	v.SetDefault("request.web", false)

	return v
}

func bindFlags(vp *viper.Viper, fs *pflag.FlagSet) {
	kv := map[string]string{
		"default.protoPath":   "path",
		"default.package":     "package",
		"default.service":     "service",
		"server.host":         "host",
		"server.port":         "port",
		"server.reflection":   "reflection",
		"server.tls":          "tls",
		"request.header":      "header",
		"request.web":         "web",
		"repl.showSplashText": "silent",
	}
	for k, v := range kv {
		f := fs.Lookup(v)
		if f == nil {
			logger.Printf("flag is not found: %s-%s", k, v)
			continue
		}
		// TODO: cleanup this special case.
		if k == "repl.showSplashText" {
			// If --silent is disabled, showSplashText is true.
			vp.Set(k, f.Value.String() != "true")
			continue
		}

		switch f.Value.Type() {
		case "stringToString":
			// There is pflag.StringToString which converts 'key=val' to a map structure.
			// However, currently, we don't use BindPFlag because it has some bugs.
			currentMap := vp.GetStringMapString(k)
			newMap := stringToStringToMap(f.Value.String())
			for k, v := range newMap {
				currentMap[k] = v
			}
			vp.Set(k, currentMap)
			continue
		case "stringSlice":
			// We want to append flag values to the config.
			// So, we don't use BindPFlag.
			currentSlice := vp.GetStringSlice(k)
			newSlice := stringSliceToSlice(f.Value.String())
			for _, v := range newSlice {
				currentSlice = append(currentSlice, v)
			}
			vp.Set(k, currentSlice)
			continue
		}
		vp.BindPFlag(k, f)
	}
}

// stringToStringToMap converts (pflag.stringToStringValue).String() to a map.
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

// writeLatestDefaultConfig writes the latest default config to path.
// Note that writeLatestDefaultConfig initializes viper again.
// So, all flags you bind by BindPFlag, global and local config will be clear.
func writeLatestDefaultConfig(path string) (*Config, error) {
	// TODO: create a new one instead of package global.
	v := newDefaultViper()
	// TODO: write tests
	// Set configVersion to the latest version.
	v.Set("meta.configVersion", meta.Version.String())

	var cfg Config
	err := v.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal default config")
	}
	setupConfig(&cfg)
	if err := v.WriteConfigAs(path); err != nil {
		return nil, errors.Wrapf(err, "failed to write the latest default config to %s", path)
	}
	return &cfg, nil
}

func initConfig(fs *pflag.FlagSet) (cfg *Config, err error) {
	v := newDefaultViper()

	defer func() {
		if fs == nil {
			logger.Println("flagset is not found")
		} else {
			logger.Println("bind flagset to the loaded config")
			bindFlags(v, fs)
			if err = v.Unmarshal(cfg); err != nil {
				return
			}
		}

		if err == nil {
			setupConfig(cfg)
		}
	}()

	cfgDir := filepath.Join(xdgbasedir.ConfigHome(), "evans")

	// Global config paths
	v.SetConfigType("toml")
	v.SetConfigName("config")
	v.AddConfigPath(cfgDir)

	logger.Printf("load global config from %s", cfgDir)
	err = v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			path := filepath.Join(cfgDir, globalConfigName)
			logger.Printf("global config is not found, create a new one: %s", path)
			if err := os.MkdirAll(cfgDir, 0755); err != nil {
				return nil, errors.Wrap(err, "failed to create config dirs")
			}
			cfg, err = writeLatestDefaultConfig(path)
			return
		}
		return nil, err
	}

	// Migrate old versions to the latest.
	if old := v.GetString("meta.configVersion"); old != meta.Version.String() {
		migrate(old, v)
		// Update the global config with the migrated config.
		logger.Println("migrated the global config to the structure of the latest version")
		v.WriteConfig()
	}

	var globalCfg Config
	if err := v.Unmarshal(&globalCfg); err != nil {
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
	if err := v.MergeConfig(f); err != nil {
		return nil, errors.Wrap(err, "failed to merge a local config to the global config")
	}

	var mergedCfg Config
	if err := v.Unmarshal(&mergedCfg); err != nil {
		return nil, err
	}
	return &mergedCfg, nil
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
}

func Edit() error {
	p, found := getLocalConfigPath()
	if !found {
		logger.Println("local config is not found. create a new local config to the project root.")
		root, found := lookupProjectRootPath()
		if !found {
			return errors.New("--edit must be call inside a Git project")
		}
		p = filepath.Join(root, localConfigName)
		logger.Printf("create a new local config to %s", p)
		if _, err := writeLatestDefaultConfig(p); err != nil {
			return err
		}
	}
	editor := getEditor()
	if editor == "" {
		return errors.New("--edit requires one of $EDITOR value or Vim")
	}
	return runEditor(editor, p)
}

var runEditor = func(editor string, cfgPath string) error {
	cmd := exec.Command(editor, cfgPath)
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
	return p, true
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
