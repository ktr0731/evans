package usecase

import (
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoprint"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/reflect/protodesc"
)

// FormatDescriptor formats the descriptor of the passed symbol.
func FormatDescriptor(symbol string) (string, error) {
	return dm.FormatDescriptor(symbol)
}
func (m *dependencyManager) FormatDescriptor(symbol string) (string, error) {
	d, err := m.descSource.FindSymbol(symbol)
	if err != nil {
		return "", errors.Wrapf(err, "failed to resolve symbol '%s'", symbol)
	}

	fd, err := desc.CreateFileDescriptor(protodesc.ToFileDescriptorProto(d.ParentFile()))
	if err != nil {
		return "", err
	}

	jd := fd.FindSymbol(symbol)
	if jd == nil {
		return "", errors.New("symbol should be found")
	}

	p := &protoprint.Printer{
		Compact:                  true,
		ForceFullyQualifiedNames: true,
		SortElements:             true,
	}
	str, err := p.PrintProtoToString(jd)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert the descriptor to string")
	}

	out := strings.TrimSpace(str)

	return fmt.Sprintf("%s:\n%s", symbol, out), nil
}
