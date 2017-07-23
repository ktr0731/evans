package config

type Core struct {
	Port string `default:"50051"`
}

type Meta struct {
	Path   string `default:"$HOME/.config"`
	Name   string `default:"evans.toml"`
	Prompt string `default:"{package}.{sevice}@{addr}:{port}"`
}

type Config struct {
	Meta Meta
	Core Core
}

var config *Config

func Load() (*Config, error) {
	return config, nil
}
