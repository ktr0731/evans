package helper

import "github.com/ktr0731/evans/config"

func TestConfig() *config.Config {
	return config.Get()
}
