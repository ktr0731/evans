package usecase

import "github.com/ktr0731/evans/grpc"
import "github.com/pkg/errors"

// ListRPCs lists all RPC belong to the selected service.
// If svcName is empty, the currently selected service will be used.
// In this case, ListRPCs doesn't modify the currently selected service.
//
// ListRPCs may return these errors:
//
//   -
func ListRPCs(svcName string) ([]*grpc.RPC, error) {
	return dm.ListRPCs(svcName)
}
func (m *dependencyManager) ListRPCs(svcName string) ([]*grpc.RPC, error) {
	if svcName == "" {
		svcName = m.state.selectedService
	}
	return m.listRPCs(m.state.selectedPackage, svcName)
}

func (m *dependencyManager) listRPCs(pkgName, svcName string) ([]*grpc.RPC, error) {
	rpcs, err := m.spec.RPCs(pkgName, svcName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list RPCs")
	}
	return rpcs, nil
}
