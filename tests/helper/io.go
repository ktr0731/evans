package helper

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func ReadAll(t *testing.T, r io.Reader) []byte {
	b, err := ioutil.ReadAll(r)
	require.NoError(t, err)
	return b
}

func ReadAllAsStr(t *testing.T, r io.Reader) string {
	return string(ReadAll(t, r))
}
