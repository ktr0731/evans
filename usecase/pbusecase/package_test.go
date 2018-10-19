package pbusecase

import (
	"testing"

	"github.com/ktr0731/evans/tests/mock/entity/mockenv"
	"github.com/ktr0731/evans/usecase/internal/usecasetest"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackage(t *testing.T) {
	expected := "example_package"
	params := &port.PackageParams{PkgName: expected}
	presenter := usecasetest.NewPresenter()
	env := &mockenv.EnvironmentMock{
		UsePackageFunc: func(string) error { return nil },
	}

	_, err := Package(params, presenter, env)
	require.NoError(t, err)
	require.Len(t, env.UsePackageCalls(), 1)
	assert.Equal(t, expected, env.UsePackageCalls()[0].Name)
}
