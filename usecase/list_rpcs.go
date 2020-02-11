package usecase

import (
	"github.com/ktr0731/evans/grpc"
	"github.com/ktr0731/evans/idl/proto"
	"github.com/pkg/errors"
)

// ListRPCs lists all RPC belong to the selected service.
// If svcName is empty, the currently selected service will be used.
// In this case, ListRPCs doesn't modify the currently selected service.
func ListRPCs(svcName string) ([]*grpc.RPC, error) {
	return dm.ListRPCs(svcName)
}
func (m *dependencyManager) ListRPCs(svcName string) ([]*grpc.RPC, error) {
	if svcName == "" {
		svcName = m.state.selectedService
	}
	fqsn := proto.FullyQualifiedServiceName(m.state.selectedPackage, svcName)
	return m.listRPCs(fqsn)
}

func (m *dependencyManager) listRPCs(fqsn string) ([]*grpc.RPC, error) {
	rpcs, err := m.spec.RPCs(fqsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list RPCs")
	}
	return rpcs, nil
}
