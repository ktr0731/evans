package protoparser

import (
	"path/filepath"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
)

func ParseFile(fnames []string, paths []string) ([]*desc.FileDescriptor, error) {
	encountered := map[string]bool{}
	paths = append(paths, ".")
	files := []string{}
	encountered["."] = true
	for _, path := range paths {
		encountered[path] = true
	}

	for _, fname := range fnames {
		_, file := filepath.Split(fname)
		files = append(files, file)

		// This is different from the result of filepath.Split above
		// because it returns '.' if the path is empty.
		path := filepath.Dir(fname)
		if !encountered[path] {
			paths = append(paths, path)
			encountered[path] = true
		}
	}
	p := &protoparse.Parser{
		ImportPaths: paths,
	}
	return p.ParseFiles(files...)
}
