package usecase

import (
	"github.com/ktr0731/evans/idl/proto"
	"github.com/pkg/errors"
)

// GetTypeDescriptor gets the descriptor of a type which belongs to the currently selected package.
func GetTypeDescriptor(typeName string) (interface{}, error) {
	return dm.GetTypeDescriptor(typeName)
}
func (m *dependencyManager) GetTypeDescriptor(typeName string) (interface{}, error) {
	pkgName := m.state.selectedPackage
	if pkgName == "" {
		pkgName = "''"
	}
	fqmn := proto.FullyQualifiedMessageName(pkgName, typeName)
	d, err := m.spec.ResolveSymbol(fqmn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get the type descriptor of '%s'", typeName)
	}
	return d, nil
}
