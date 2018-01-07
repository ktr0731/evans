package gateway

import (
	"path/filepath"
	"testing"

	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_fqrnToEndpoint(t *testing.T) {
	env := setupEnv(t, filepath.Join("helloworld", "helloworld.proto"), "helloworld", "Greeter")
	rpc, err := env.RPC("SayHello")
	require.NoError(t, err)

	// TODO: don't ignore connection error log
	client, err := NewGRPCClient(helper.TestConfig())
	require.NoError(t, err)

	fqrn, err := client.fqrnToEndpoint(rpc.FQRN)
	require.NoError(t, err)
	assert.Equal(t, fqrn, "/helloworld.Greeter/SayHello")
}
