package usecase

import (
	"testing"
)

func TestGetPreviousRPCRequestReturnsErrorWhenPreviousDoesNotExist(t *testing.T) {
	d := &dependencyManager{
		state: state{
			rpcCallState: nil,
		},
	}
	var req interface{}
	if err := d.getPreviousRPCRequest("unknown", ""); err == nil {
		t.Errorf("expected an error but got %s", req)
	}
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
	var req interface{}
	if err := d.getPreviousRPCRequest("TestRPC", ""); err == nil {
		t.Errorf("expected an error but got %s", req)
	}
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
	var req interface{}
	if err := d.getPreviousRPCRequest("TestRPC", ""); err == nil {
		t.Errorf("expected an error but got %s", req)
	}
}
