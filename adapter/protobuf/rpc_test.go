package protobuf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRPC(t *testing.T) {
	d := parseFile(t, []string{"helloworld.proto"}, nil)
	require.Len(t, d, 1)

	svcs := d[0].GetServices()
	require.Len(t, svcs, 1)

	msgs := d[0].GetMessageTypes()
	require.Len(t, msgs, 2)

	reqMsg := msgs[0]
	resMsg := msgs[1]

	svc := newService(svcs[0])
	require.Len(t, svc.RPCs(), 1)

	rpc := svc.RPCs()[0]

	require.Equal(t, "SayHello", rpc.Name())
	require.Equal(t, "helloworld.Greeter.SayHello", rpc.FQRN())
	require.Equal(t, reqMsg.GetName(), rpc.RequestMessage().Name())
	require.Equal(t, len(reqMsg.GetFields()), len(rpc.RequestMessage().Fields()))
	require.Equal(t, resMsg.GetName(), rpc.ResponseMessage().Name())
	require.Equal(t, len(resMsg.GetFields()), len(rpc.ResponseMessage().Fields()))
}
