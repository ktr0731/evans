package protoparser

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
)

// ParseFile parses proto files to []*desc.FileDescriptor
func ParseFile(filename []string, paths []string) ([]*desc.FileDescriptor, error) {
	p := &protoparse.Parser{
		ImportPaths: append(paths, "."),
	}
	return p.ParseFiles(filename...)
}
