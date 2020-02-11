package usecase

import (
	"github.com/pkg/errors"
)

// FormatMethod formats a method.
func FormatMethod(fqmn string) (string, error) {
	return dm.FormatMethod(fqmn)
}
func (m *dependencyManager) FormatMethod(fqmn string) (string, error) {
	fqsn, _, err := ParseFullyQualifiedMethodName(fqmn)
	if err != nil {
		return "", err
	}
	v, err := m.methodsToFormatStructs(fqsn)
	if err != nil {
		return "", err
	}
	for _, method := range v.Methods {
		if fqmn != method.FullyQualifiedName {
			continue
		}
		out, err := m.resourcePresenter.Format(method)
		if err != nil {
			return "", errors.Wrap(err, "failed to format method names by presenter")
		}
		return out, nil
	}
	return "", errors.New("method is not found")
}
