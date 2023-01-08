package usecase

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestGetPreviousRPCRequest(t *testing.T) {
	cases := map[string]struct {
		expectedError string
		method        protoreflect.MethodDescriptor
		rpcCallState  map[rpcIdentifier]callState
	}{
		"no previous request exists": {
			method:        getStubMethod(false),
			expectedError: "no previous request exists for method: TestRPC, please issue a normal request",
		},
		"previous request is client streaming": {
			rpcCallState:  map[rpcIdentifier]callState{"TestRPC": {}},
			method:        getStubMethod(true),
			expectedError: "cannot rerun previous method: TestRPC as client/bidi streaming RPCs are not supported",
		},
		"previous request bytes are nil": {
			rpcCallState:  map[rpcIdentifier]callState{"TestRPC": {}},
			method:        getStubMethod(false),
			expectedError: "no previous request body exists for method: TestRPC, please issue a normal request",
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
			var req proto.Message
			err := d.getPreviousRPCRequest(c.method, req)
			if err == nil || err.Error() != c.expectedError {
				t.Errorf("expected error %s, but got %s", c.expectedError, err)
			}
		})
	}
}

type stubMethod struct {
	protoreflect.MethodDescriptor

	isStreamingClient bool
}

func (m *stubMethod) FullName() protoreflect.FullName { return protoreflect.FullName("TestRPC") }
func (m *stubMethod) IsStreamingClient() bool         { return m.isStreamingClient }
func (m *stubMethod) IsStreamingServer() bool         { return true }

func getStubMethod(clientStreaming bool) *stubMethod {
	return &stubMethod{
		isStreamingClient: clientStreaming,
	}
}
