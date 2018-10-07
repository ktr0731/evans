package usecase

import (
	"testing"

	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/entity/env"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type headerEnv struct {
	env.Environment

	removeCalled bool
}

func (e *headerEnv) AddHeader(_ *entity.Header) {}

func (e *headerEnv) RemoveHeader(_ string) {
	e.removeCalled = true
}

func TestHeader(t *testing.T) {
	params := &port.HeaderParams{
		Headers: []*entity.Header{
			{Key: "tsukasa", Val: "ayatsuji"},
			{Key: "miya", Val: "tachibana", NeedToRemove: true},
		},
	}
	presenter := presenter.NewStubPresenter()
	env := &headerEnv{}

	r, err := Header(params, presenter, env)
	require.NoError(t, err)
	assert.Equal(t, r, nil)

	assert.True(t, env.removeCalled)
}
