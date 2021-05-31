package usecase

import (
	"testing"

	"github.com/ktr0731/evans/grpc"

	"github.com/jhump/protoreflect/dynamic"
)

func TestGetPreviousRPCRequest(t *testing.T) {
	cases := map[string]struct {
		expectedError string
		rpc           *grpc.RPC
		rpcCallState  map[rpcIdentifier]callState
	}{
		"no previous request exists": {
			rpc:           getStubRPC(false),
			expectedError: "no previous request exists for RPC: TestRPC, please issue a normal request",
		},
		"previous request is client streaming": {
			rpcCallState:  map[rpcIdentifier]callState{"TestRPC": {}},
			rpc:           getStubRPC(true),
			expectedError: "cannot rerun previous RPC: TestRPC as client/bidi streaming RPCs are not supported",
		},
		"previous request bytes are nil": {
			rpcCallState:  map[rpcIdentifier]callState{"TestRPC": {}},
			rpc:           getStubRPC(false),
			expectedError: "no previous request body exists for RPC: TestRPC, please issue a normal request",
		},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			d := &dependencyManager{
				state: state{
					rpcCallState: c.rpcCallState,
				},
			}
			var req *dynamic.Message
			err := d.getPreviousRPCRequest(c.rpc, req)
			if err == nil || err.Error() != c.expectedError {
				t.Errorf("expected error %s, but got %s", c.expectedError, err)
			}
		})
	}
}

func getStubRPC(clientStreaming bool) *grpc.RPC {
	return &grpc.RPC{
		Name:              "TestRPC",
		IsServerStreaming: true,
		IsClientStreaming: clientStreaming,
	}
}
