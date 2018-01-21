package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPC(t *testing.T) {
	d := parseFile(t, "helloworld.proto")
	svcs := d.GetServices()
	require.Len(t, svcs, 1)

	msgs := d.GetMessageTypes()
	require.Len(t, msgs, 2)

	reqMsg := msgs[0]
	resMsg := msgs[1]

	svc := newService(svcs[0])
	assert.Len(t, svc.RPCs, 1)

	rpc := svc.RPCs[0]

	assert.Equal(t, "SayHello", rpc.Name)
	assert.Equal(t, "helloworld.Greeter.SayHello", rpc.FQRN)
	assert.Equal(t, reqMsg, rpc.RequestType)
	assert.Equal(t, resMsg, rpc.ResponseType)
}
