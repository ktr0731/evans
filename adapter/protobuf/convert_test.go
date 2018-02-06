package protobuf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToEntitiesFrom(t *testing.T) {
	d := parseDependFiles(t, "helloworld.proto")
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
