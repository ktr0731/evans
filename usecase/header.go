package usecase

import (
	"strings"

	"github.com/ktr0731/evans/grpc"
	"github.com/ktr0731/evans/logger"
)

func AddHeader(k, v string) {
	dm.AddHeader(k, v)
}
func (m *dependencyManager) AddHeader(k, v string) {
	if strings.ToLower(k) == "user-agent" {
		logger.Println(`warning: cannot add a header named "user-agent"`)
		return
	}
	if err := m.gRPCClient.Headers().Add(k, v); err != nil {
		logger.Printf("failed to add a header %s=%s: %s", k, v, err)
	}
}

func RemoveHeader(k string) {
	dm.RemoveHeader(k)
}
func (m *dependencyManager) RemoveHeader(k string) {
	m.gRPCClient.Headers().Remove(k)
}

func ListHeaders() grpc.Headers {
	return dm.ListHeaders()
}
func (m *dependencyManager) ListHeaders() grpc.Headers {
	return m.gRPCClient.Headers()
}
