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
	m.state.headers.Add(k, v)
}

func RemoveHeader(k string) {
	dm.RemoveHeader(k)
}
func (m *dependencyManager) RemoveHeader(k string) {
	m.state.headers.Remove(k)
}

func ListHeaders() grpc.Headers {
	return dm.ListHeaders()
}
func (m *dependencyManager) ListHeaders() grpc.Headers {
	return m.state.headers
}
