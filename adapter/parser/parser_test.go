package parser

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile(t *testing.T) {
	t.Run("importing", func(t *testing.T) {
		fnames := []string{testdata("importing", "library.proto")}
		pkgs, err := ParseFile(fnames, nil)
		require.NoError(t, err)
		assert.Len(t, pkgs, 1)
		assert.Len(t, pkgs[0].Messages, 4)
	})
}

func testdata(s ...string) string {
	return filepath.Join(append([]string{"testdata"}, s...)...)
}
