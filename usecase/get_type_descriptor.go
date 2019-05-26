package usecase

import "github.com/pkg/errors"

// GetTypeDescriptor gets the descriptor of a type which belongs to the currently selected package.
func GetTypeDescriptor(typeName string) (interface{}, error) {
	return dm.GetTypeDescriptor(typeName)
}
func (m *dependencyManager) GetTypeDescriptor(typeName string) (interface{}, error) {
	d, err := m.spec.TypeDescriptor(m.state.selectedPackage, typeName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get the type descriptor of '%s'", typeName)
	}
	return d, nil
}
