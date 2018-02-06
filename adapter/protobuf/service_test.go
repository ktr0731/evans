package protobuf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	d := parseFile(t, "helloworld.proto")
	svcs := d.GetServices()
	require.Len(t, svcs, 1)

	svc := newService(svcs[0])
	require.Equal(t, "Greeter", svc.Name())
	require.Len(t, svc.RPCs(), 1)
}
