package proto

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type DescriptorSource interface {
	FindSymbol(name string) (protoreflect.MessageDescriptor, error)
}

type reflection struct {
	client messageTypeRegistry
}

type messageTypeRegistry interface {
	FindSymbol(name string) (protoreflect.MessageDescriptor, error)
}

func NewDescriptorSourceFromReflection(c messageTypeRegistry) DescriptorSource {
	return &reflection{c}
}

func (r *reflection) FindSymbol(name string) (protoreflect.MessageDescriptor, error) {
	return r.client.FindSymbol(name)
}

type files struct {
	fds []*desc.FileDescriptor
}

func NewDescriptorSourceFromFiles(importPaths []string, fnames []string) (DescriptorSource, error) {
	p := &protoparse.Parser{
		ImportPaths: importPaths,
	}
	fds, err := p.ParseFiles(fnames...)
	if err != nil {
		return nil, errors.Wrap(err, "proto: failed to parse passed proto files")
	}

	return &files{fds: fds}, nil
}

var errSymbolNotFound = errors.New("proto: symbol not found")

func (f *files) FindSymbol(name string) (protoreflect.MessageDescriptor, error) {
	for _, fd := range f.fds {
		d := fd.FindSymbol(name)
		if d == nil {
			continue
		}

		if err := RegisterFileAndType(fd); err != nil {
			return nil, errors.Wrap(err, "failed to register file dscriptor")
		}

		fd, err := protodesc.NewFile(fd.AsFileDescriptorProto(), protoregistry.GlobalFiles)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to find file containing symbol %s", name)
		}

		md := fd.Messages().ByName(protoreflect.FullName(name).Name())
		if md == nil {
			return nil, fmt.Errorf("failed to find message '%s'", name)
		}
	}

	return nil, errSymbolNotFound
}
