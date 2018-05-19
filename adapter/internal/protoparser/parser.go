package protoparser

import (
	"path/filepath"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
)

// ParseFile parses proto files to []*desc.FileDescriptor
func ParseFile(fnames []string, paths []string) ([]*desc.FileDescriptor, error) {
	encountered := map[string]bool{}
	paths = append(paths, ".")
	encountered["."] = true
	for _, path := range paths {
		encountered[path] = true
	}

	for _, fname := range fnames {
		path := filepath.Dir(fname)
		if !encountered[path] {
			paths = append(paths, path)
			encountered[path] = true
		}
	}
	p := &protoparse.Parser{
		ImportPaths: paths,
	}
	return p.ParseFiles(fnames...)
}
