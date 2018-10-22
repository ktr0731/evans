package grpc

import (
	"testing"

	"github.com/ktr0731/evans/adapter/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func Test_fqrnToEndpoint(t *testing.T) {
	env := testhelper.SetupEnv(t, "helloworld.proto", "helloworld", "Greeter")
	rpc, err := env.RPC("SayHello")
	require.NoError(t, err)

	fqrn, err := fqrnToEndpoint(rpc.FQRN())
	require.NoError(t, err)
	require.Equal(t, fqrn, "/helloworld.Greeter/SayHello")
}
