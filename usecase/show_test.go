package usecase

import (
	"testing"

	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type showEnv struct {
	entity.Environment

	expected entity.Packages
}

func (e *showEnv) Packages() entity.Packages {
	return e.expected
}

func TestShow(t *testing.T) {
	expected := entity.Packages{{Name: "example_package"}}
	params := &port.ShowParams{Type: port.ShowTypePackage}
	presenter := presenter.NewJSONCLIPresenter()
	env := &showEnv{expected: expected}

	res, err := Show(params, presenter, env)
	require.NoError(t, err)

	actual := helper.ReadAllAsStr(t, res)

	assert.Equal(t, packages(expected).Show(), actual)
}
