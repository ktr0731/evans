package protobuf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile(t *testing.T) {
	t.Run("importing", func(t *testing.T) {
		fnames := []string{"library.proto"}
		pkgs, err := ParseFile(fnames, []string{"testdata/importing"})
		require.NoError(t, err)
		assert.Len(t, pkgs, 1)
		assert.Len(t, pkgs[0].Messages, 4)
	})
}
