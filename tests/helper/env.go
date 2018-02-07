package helper

import (
	"testing"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/stretchr/testify/require"
)

func NewEnv(t *testing.T, desc []*entity.Package, config *config.Config) *entity.Env {
	env, err := entity.NewEnv(desc, config)
	require.NoError(t, err)
	return env
}
