package usecase

import (
	"testing"

	"github.com/jhump/protoreflect/dynamic"

	"github.com/stretchr/testify/assert"
)

func TestGetPreviousRPCRequestReturnsErrorWhenPreviousDoesNotExist(t *testing.T) {
	d := &dependencyManager{
		state: state{
			rpcCallState: nil,
		},
	}
	var req *dynamic.Message
	err := d.getPreviousRPCRequest("TestRPC", req)
	assert.Error(t, err)
	assert.Equal(t, "no previous request exists for RPC: TestRPC, please issue a normal request", err.Error())
}

func TestGetPreviousRPCRequestReturnsErrorWhenRequestIsNotRepeatable(t *testing.T) {
	d := &dependencyManager{
		state: state{
			rpcCallState: map[rpcIdentifier]callState{
				"TestRPC": {
					repeatable: false,
				},
			},
		},
	}
	var req *dynamic.Message
	err := d.getPreviousRPCRequest("TestRPC", req)
	assert.Error(t, err)
	assert.Equal(t, "cannot rerun previous RPC: TestRPC as client/bidi streaming RPCs are not supported", err.Error())
}

func TestGetPreviousRPCRequestReturnsErrorWhenRequestBytesAreNil(t *testing.T) {
	d := &dependencyManager{
		state: state{
			rpcCallState: map[rpcIdentifier]callState{
				"TestRPC": {
					repeatable: true,
					reqPayload: nil,
				},
			},
		},
	}
	var req *dynamic.Message
	err := d.getPreviousRPCRequest("TestRPC", req)
	assert.Error(t, err)
	assert.Equal(t, "no previous request body exists for RPC: TestRPC, please issue a normal request", err.Error())
}
