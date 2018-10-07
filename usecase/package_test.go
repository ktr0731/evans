package usecase

import (
	"testing"

	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity/env"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/require"
)

type packageEnv struct {
	env.Environment

	usedPackage string
}

func (e *packageEnv) UsePackage(pkgName string) error {
	e.usedPackage = pkgName
	return nil
}

func TestPackage(t *testing.T) {
	expected := "example_package"
	params := &port.PackageParams{PkgName: expected}
	presenter := &presenter.StubPresenter{}
	env := &packageEnv{}

	_, err := Package(params, presenter, env)
	require.NoError(t, err)
	require.Equal(t, expected, env.usedPackage)
}
