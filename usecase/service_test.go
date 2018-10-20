package usecase

import (
	"testing"

	"github.com/ktr0731/evans/tests/mock/entity/mockenv"
	"github.com/ktr0731/evans/usecase/internal/usecasetest"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	expected := "example_service"
	params := &port.ServiceParams{SvcName: expected}
	presenter := usecasetest.NewPresenter()
	env := &mockenv.EnvironmentMock{
		UseServiceFunc: func(string) error { return nil },
	}

	_, err := Service(params, presenter, env)
	require.NoError(t, err)

	assert.Len(t, env.UseServiceCalls(), 1)
	assert.Equal(t, expected, env.UseServiceCalls()[0].Name)
}
