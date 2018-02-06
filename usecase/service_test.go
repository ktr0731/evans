package usecase

import (
	"testing"

	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/require"
)

type serviceEnv struct {
	entity.Environment

	usedService string
}

func (e *serviceEnv) UseService(pkgName string) error {
	e.usedService = pkgName
	return nil
}

func TestService(t *testing.T) {
	expected := "example_service"
	params := &port.ServiceParams{expected}
	presenter := &presenter.StubPresenter{}
	env := &serviceEnv{}

	_, err := Service(params, presenter, env)
	require.NoError(t, err)
	require.Equal(t, expected, env.usedService)
}
