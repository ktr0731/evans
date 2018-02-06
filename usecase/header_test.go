package usecase

import (
	"errors"
	"testing"

	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/require"
)

type headerEnv struct {
	entity.Environment
	err error
}

func (e *headerEnv) AddHeader(_ *entity.Header) error {
	return e.err
}

func (e *headerEnv) RemoveHeader(_ string) {}

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
	require.Equal(t, r, nil)

	env.err = errors.New("an error")
	r, err = Header(params, presenter, env)
	require.Error(t, err)
}
