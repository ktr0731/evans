package usecase

import (
	"strings"

	"github.com/ktr0731/evans/proto"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GetTypeDescriptor gets the descriptor of a type which belongs to the currently selected package.
func GetTypeDescriptor(typeName string) (protoreflect.Descriptor, error) {
	return dm.GetTypeDescriptor(typeName)
}
func (m *dependencyManager) GetTypeDescriptor(typeName string) (protoreflect.Descriptor, error) {
	pkgName := m.state.selectedPackage

	fqmn := typeName
	if !strings.HasPrefix(typeName, pkgName+".") {
		fqmn = proto.FullyQualifiedMessageName(pkgName, typeName)
	}

	d, err := m.descSource.FindSymbol(fqmn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get the type descriptor of '%s'", typeName)
	}
	return d, nil
}
