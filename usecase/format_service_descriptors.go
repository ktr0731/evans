package usecase

import (
	"strings"

	"github.com/pkg/errors"
)

// FormatServiceDescriptors formats all service descriptors the spec loaded.
func FormatServiceDescriptors() (string, error) {
	return dm.FormatServiceDescriptors()
}
func (m *dependencyManager) FormatServiceDescriptors() (string, error) {
	svcs, err := m.ListServices()
	if err != nil {
		return "", err
	}

	out := make([]string, 0, len(svcs))
	for _, s := range svcs {
		o, err := FormatDescriptor(s)
		if err != nil {
			return "", errors.Wrap(err, "failed to format one service descriptor")
		}
		out = append(out, o)
	}
	return strings.Join(out, "\n\n"), nil
}
