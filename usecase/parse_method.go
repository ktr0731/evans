package usecase

import (
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	v, err := m.descSource.FindSymbol(fqmn)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to find the symbol")
	}
	if _, ok := v.(protoreflect.MethodDescriptor); !ok {
		return "", "", errors.New("symbol is not method descriptor")
	}

	svc, mtd := fqmn[:i], fqmn[i+1:]
	return svc, mtd, nil
}
