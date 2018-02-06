package protobuf

import (
	"testing"

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
	require.Len(t, svc.RPCs, 1)

	rpc := svc.RPCs()[0]

	require.Equal(t, "SayHello", rpc.Name)
	require.Equal(t, "helloworld.Greeter.SayHello", rpc.FQRN)
	require.Equal(t, reqMsg, rpc.RequestMessage())
	require.Equal(t, resMsg, rpc.ResponseMessage())
}
