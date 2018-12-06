package protobuf

import (
	"path/filepath"
	"testing"

	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/adapter/internal/protoparser"
	"github.com/stretchr/testify/require"
)

func parseFile(t *testing.T, fnames []string, paths []string) []*desc.FileDescriptor {
	for i := range fnames {
		fnames[i] = filepath.Join("testdata", fnames[i])
	}
	d, err := protoparser.ParseFile(fnames, paths)
	require.NoError(t, err)
	return d
}
