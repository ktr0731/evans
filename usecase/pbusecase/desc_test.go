package pbusecase

import (
	"testing"

	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/entity/testentity"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/ktr0731/evans/tests/mock/entity/mockenv"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/require"
)

func TestDescribe(t *testing.T) {
	var expected entity.Message = testentity.NewMsg()

	params := &port.DescribeParams{}
	presenter := presenter.NewJSONCLIPresenter()
	env := &mockenv.EnvironmentMock{
		MessageFunc: func(name string) (entity.Message, error) { return expected, nil },
	}

	res, err := Describe(params, presenter, env)
	require.NoError(t, err)

	actual := helper.ReadAllAsStr(t, res)
	m := &message{expected}
	require.Equal(t, m.Show(), actual)
}
