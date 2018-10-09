package usecase

import (
	"testing"

	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity/env"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	expected := "example_service"
	params := &port.ServiceParams{SvcName: expected}
	presenter := &presenter.StubPresenter{}
	env := &env.EnvironmentMock{
		UseServiceFunc: func(string) error { return nil },
	}

	_, err := Service(params, presenter, env)
	require.NoError(t, err)

	require.Len(t, env.UseServiceCalls(), 1)
	require.Equal(t, expected, env.UseServiceCalls()[0].Name)
}
