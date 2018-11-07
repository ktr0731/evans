package protobuf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToEntitiesFrom(t *testing.T) {
	d := parseFile(t, []string{"helloworld.proto"}, nil)
	p, err := ToEntitiesFrom(d)
	require.NoError(t, err)

	require.Len(t, p, 1)

	pkg := p[0]
	require.Len(t, pkg.Messages, 2)
	require.Len(t, pkg.Services, 1)
	require.Len(t, pkg.Services[0].RPCs(), 1)

	rpc := pkg.Services[0].RPCs()[0]
	require.Equal(t, rpc.RequestMessage(), pkg.Messages[0])
	require.Equal(t, rpc.ResponseMessage(), pkg.Messages[1])

	require.Len(t, pkg.Messages[0].Fields(), 2)
	require.Len(t, pkg.Messages[1].Fields(), 2)
}

func TestToEntitiesFromServiceDescriptors(t *testing.T) {
	d := parseFile(t, []string{"helloworld.proto"}, nil)
	require.Len(t, d, 1)

	svcs, msgs := ToEntitiesFromServiceDescriptors(d[0].GetServices())
	assert.Len(t, svcs, len(d))
	assert.Lenf(t, msgs, 2, "expected number of messages is same as number of request/response message, but got %d", len(msgs))

	assert.Len(t, svcs[0].RPCs(), 1)
}
