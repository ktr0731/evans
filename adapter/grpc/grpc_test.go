package grpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_fqrnToEndpoint(t *testing.T) {
	fqrn := "helloworld.Greeter.SayHello"

	endpoint, err := fqrnToEndpoint(fqrn)
	require.NoError(t, err)
	require.Equal(t, endpoint, "/helloworld.Greeter/SayHello")
}
