package helper

import (
	"path/filepath"
	"testing"

	"github.com/ktr0731/evans/adapter/parser"
	"github.com/ktr0731/evans/entity"
	"github.com/stretchr/testify/require"
)

func ReadProto(t *testing.T, fpath ...string) []*entity.Package {
	for i := range fpath {
		fpath[i] = filepath.Join("testdata", fpath[i])
	}
	set, err := parser.ParseFile(fpath, nil)
	require.NoError(t, err)
	return set
}
