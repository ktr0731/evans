package usecase

import (
	"errors"
	"strings"
)

// ParseFullyQualifiedMethodName parses the passed fully-qualified method as fully-qualified service name and method name.
// ParseFullyQualifiedMethodName may return these errors:
//
//   - An error described in idl.Spec.RPC method returns.
//   - An error if fqmn is not a valid fully-qualified method name form.
//
func ParseFullyQualifiedMethodName(fqmn string) (fqsn, method string, err error) {
	return dm.ParseFullyQualifiedMethodName(fqmn)
}
func (m *dependencyManager) ParseFullyQualifiedMethodName(fqmn string) (string, string, error) {
	i := strings.LastIndex(fqmn, ".")
	if i == -1 {
		return "", "", errors.New("invalid fully-qualified method name")
	}
	svc, mtd := fqmn[:i], fqmn[i+1:]
	_, err := m.spec.RPC(svc, mtd)
	return svc, mtd, err
}
