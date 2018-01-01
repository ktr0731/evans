package helper

import (
	"testing"

	"github.com/ktr0731/evans/parser"
	"github.com/stretchr/testify/require"
)

const (
	protoPath = "tests/testdata"
)

func ReadProto(t *testing.T, filepath []string) *parser.FileDescriptorSet {
	set, err := parser.ParseFile(filepath, []string{protoPath})
	require.NoError(t, err)
	return set
}
