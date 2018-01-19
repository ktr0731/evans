package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	d := parseFile(t, "helloworld.proto")
	svcs := d.GetServices()
	require.Len(t, svcs, 1)

	svc := newService(svcs[0])
	assert.Equal(t, "Greeter", svc.Name)
	assert.Len(t, svc.RPCs, 1)
}
