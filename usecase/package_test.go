package usecase

import (
	"testing"

	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type packageEnv struct {
	entity.Environment

	usedPackage string
}

func (e *packageEnv) UsePackage(pkgName string) error {
	e.usedPackage = pkgName
	return nil
}

func TestPackage(t *testing.T) {
	expected := "example_package"
	params := &port.PackageParams{expected}
	presenter := &presenter.StubPresenter{}
	env := &packageEnv{}

	_, err := Package(params, presenter, env)
	require.NoError(t, err)
	assert.Equal(t, expected, env.usedPackage)
}
