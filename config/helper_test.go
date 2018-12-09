package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func DefaultConfig(t *testing.T) *Config {
	cfg, err := defaultConfig()
	require.NoError(t, err, "Unmarshal must not return any errors")
	return cfg
}
