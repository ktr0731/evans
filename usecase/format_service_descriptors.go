package usecase

import (
	"strings"

	"github.com/pkg/errors"
)

// FormatFileDescriptors formats all file descriptors the spec loaded.
func FormatFileDescriptors() (string, error) {
	return dm.FormatFileDescriptors()
}
func (m *dependencyManager) FormatFileDescriptors() (string, error) {
	svcs := ListServices()
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
