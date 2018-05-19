package protobuf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	d := parseFile(t, []string{"helloworld.proto"}, nil)
	require.Len(t, d, 1)

	svcs := d[0].GetServices()
	require.Len(t, svcs, 1)

	svc := newService(svcs[0])
	require.Equal(t, "Greeter", svc.Name())
	require.Len(t, svc.RPCs(), 1)
}
