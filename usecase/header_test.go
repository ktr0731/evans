package usecase

import (
	"errors"
	"testing"

	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type headerEnv struct {
	entity.Environment
	err error
}

func (e *headerEnv) AddHeaders(h ...*entity.Header) error {
	return e.err
}

func TestHeader(t *testing.T) {
	params := &port.HeaderParams{
		Headers: []*entity.Header{
			{Key: "tsukasa", Val: "ayatsuji"},
		},
	}
	presenter := presenter.NewStubPresenter()
	env := &headerEnv{}

	r, err := Header(params, presenter, env)
	require.NoError(t, err)
	assert.Equal(t, r, nil)

	env.err = errors.New("an error")
	r, err = Header(params, presenter, env)
	require.Error(t, err)
}
