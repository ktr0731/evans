package usecase

import (
	"fmt"

	"github.com/pkg/errors"
)

// FormatDescriptor formats all package names.
func FormatDescriptor(symbol string) (string, error) {
	return dm.FormatDescriptor(symbol)
}
func (m *dependencyManager) FormatDescriptor(symbol string) (string, error) {
	v, err := m.spec.ResolveSymbol(symbol)
	if err != nil {
		return "", errors.Wrapf(err, "failed to resolve symbol '%s'", symbol)
	}
	out, err := m.spec.FormatDescriptor(v)
	if err != nil {
		return "", errors.Wrapf(err, "failed to format the descriptor of symbol '%s'", symbol)
	}
	return fmt.Sprintf("%s:\n%s", symbol, out), nil
}
