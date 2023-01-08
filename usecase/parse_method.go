package usecase

import (
	"strings"

	"github.com/pkg/errors"
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
	if _, err := m.descSource.FindSymbol(fqmn); err != nil {
		return "", "", errors.Wrap(err, "failed to find the symbol")
	}

	svc, mtd := fqmn[:i], fqmn[i+1:]
	return svc, mtd, nil
}
