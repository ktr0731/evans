package helper

import (
	"testing"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/parser"
	"github.com/stretchr/testify/require"
)

func NewEnv(t *testing.T, desc *parser.FileDescriptorSet, config *config.Env) *entity.Env {
	env, err := entity.New(desc, config)
	require.NoError(t, err)
	return env
}
