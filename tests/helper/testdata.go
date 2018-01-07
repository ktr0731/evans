package helper

import (
	"path/filepath"
	"testing"

	"github.com/ktr0731/evans/parser"
	"github.com/stretchr/testify/require"
)

func ReadProto(t *testing.T, fpath []string) *parser.FileDescriptorSet {
	for i := range fpath {
		fpath[i] = filepath.Join("testdata", fpath[i])
	}
	set, err := parser.ParseFile(fpath, nil)
	require.NoError(t, err)
	return set
}
