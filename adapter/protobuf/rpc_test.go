package protobuf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPC(t *testing.T) {
	d := parseFile(t, "stream.proto")
	svcs := d.GetServices()
	assert.Len(t, svcs, 1)

	msgs := d.GetMessageTypes()
	assert.Len(t, msgs, 2)

	reqMsg := msgs[0]
	resMsg := msgs[1]

	svc := newService(svcs[0])
	assert.Len(t, svc.RPCs(), 1)

	rpc := svc.RPCs()[0]

	assert.Equal(t, "SayHelloClientStreaming", rpc.Name())
	assert.Equal(t, "helloworld.Greeter.SayHelloClientStreaming", rpc.FQRN())
	assert.Equal(t, reqMsg.GetName(), rpc.RequestMessage().Name())
	assert.Equal(t, len(reqMsg.GetFields()), len(rpc.RequestMessage().Fields()))
	assert.Equal(t, resMsg.GetName(), rpc.ResponseMessage().Name())
	assert.Equal(t, len(resMsg.GetFields()), len(rpc.ResponseMessage().Fields()))
	assert.True(t, rpc.IsClientStreaming())
}
