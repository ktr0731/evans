package usecase

import (
	"testing"

	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/entity/testentity"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type describeEnv struct {
	entity.Environment

	expected entity.Message
}

func (e *describeEnv) Message(msgName string) (entity.Message, error) {
	return e.expected, nil
}

func TestDescribe(t *testing.T) {
	var expected entity.Message = testentity.NewMsg()

	params := &port.DescribeParams{}
	presenter := presenter.NewJSONCLIPresenter()
	env := &describeEnv{expected: expected}

	res, err := Describe(params, presenter, env)
	require.NoError(t, err)

	actual := helper.ReadAllAsStr(t, res)
	m := &message{expected}
	assert.Equal(t, m.Show(), actual)
}
