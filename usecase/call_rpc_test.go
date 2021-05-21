package usecase

import (
	"testing"

	"github.com/ktr0731/evans/grpc"

	"github.com/jhump/protoreflect/dynamic"

	"github.com/stretchr/testify/assert"
)

func TestGetPreviousRPCRequestReturnsErrorWhenPreviousDoesNotExist(t *testing.T) {
	d := &dependencyManager{
		state: state{
			rpcCallState: nil,
		},
	}
	rpc := getStubRPC(false)
	var req *dynamic.Message
	err := d.getPreviousRPCRequest(rpc, req)
	assert.Error(t, err)
	assert.Equal(t, "no previous request exists for RPC: TestRPC, please issue a normal request", err.Error())
}

func TestGetPreviousRPCRequestReturnsErrorWhenRequestIsOfTypeClientStreaming(t *testing.T) {
	d := &dependencyManager{
		state: state{
			rpcCallState: map[rpcIdentifier]callState{
				"TestRPC": {},
			},
		},
	}
	rpc := getStubRPC(true)
	var req *dynamic.Message
	err := d.getPreviousRPCRequest(rpc, req)
	assert.Error(t, err)
	assert.Equal(t, "cannot rerun previous RPC: TestRPC as client/bidi streaming RPCs are not supported", err.Error())
}

func TestGetPreviousRPCRequestReturnsErrorWhenRequestBytesAreNil(t *testing.T) {
	d := &dependencyManager{
		state: state{
			rpcCallState: map[rpcIdentifier]callState{
				"TestRPC": {
					requestPayload: nil,
				},
			},
		},
	}
	rpc := getStubRPC(false)
	var req *dynamic.Message
	err := d.getPreviousRPCRequest(rpc, req)
	assert.Error(t, err)
	assert.Equal(t, "no previous request body exists for RPC: TestRPC, please issue a normal request", err.Error())
}

func getStubRPC(clientStreaming bool) *grpc.RPC {
	return &grpc.RPC{
		Name:              "TestRPC",
		IsServerStreaming: true,
		IsClientStreaming: clientStreaming,
	}
}
