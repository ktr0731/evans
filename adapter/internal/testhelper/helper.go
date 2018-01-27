package testhelper

import (
	"testing"

	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/require"
)

func SetupEnv(t *testing.T, fpath, pkgName, svcName string) *entity.Env {
	t.Helper()

	set := helper.ReadProto(t, fpath)

	env := helper.NewEnv(t, set, helper.TestConfig().Env)

	err := env.UsePackage(pkgName)
	require.NoError(t, err)

	if svcName != "" {
		err = env.UseService(svcName)
		require.NoError(t, err)
	}

	return env
}
