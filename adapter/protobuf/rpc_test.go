package protobuf

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPC_stream(t *testing.T) {
	d := parseFile(t, "stream.proto")
	svcs := d.GetServices()
	assert.Len(t, svcs, 1)

	msgs := d.GetMessageTypes()
	assert.Len(t, msgs, 2)

	reqMsg := msgs[0]
	resMsg := msgs[1]
	svc := newService(svcs[0])

	cases := []struct {
		name                                 string
		isClientStreaming, isServerStreaming bool
	}{
		{name: "SayHelloUnary"},
		{name: "SayHelloClientStreaming", isClientStreaming: true},
	}

	assert.Len(t, svc.RPCs(), len(cases))

	for i, c := range cases {
		rpc := svc.RPCs()[i]
		assert.Equal(t, c.name, rpc.Name())
		assert.Equal(t, fmt.Sprintf("helloworld.Greeter.%s", c.name), rpc.FQRN())
		assert.Equal(t, reqMsg.GetName(), rpc.RequestMessage().Name())
		assert.Equal(t, len(reqMsg.GetFields()), len(rpc.RequestMessage().Fields()))
		assert.Equal(t, resMsg.GetName(), rpc.ResponseMessage().Name())
		assert.Equal(t, len(resMsg.GetFields()), len(rpc.ResponseMessage().Fields()))
		assert.Equal(t, c.isClientStreaming, rpc.IsClientStreaming())
		assert.Equal(t, c.isServerStreaming, rpc.IsServerStreaming())

		if !c.isClientStreaming && !c.isServerStreaming {
			assert.Panics(t, func() {
				rpc.StreamDesc()
			})
		} else {
			assert.NotPanics(t, func() {
				rpc.StreamDesc()
			})
		}
	}
}
