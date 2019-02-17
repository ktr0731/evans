package grpc

import (
	"testing"

	"github.com/ktr0731/evans/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_fqrnToEndpoint(t *testing.T) {
	fqrn := "helloworld.Greeter.SayHello"

	endpoint, err := fqrnToEndpoint(fqrn)
	require.NoError(t, err)
	require.Equal(t, endpoint, "/helloworld.Greeter/SayHello")
}

func TestNewClient(t *testing.T) {
	cases := map[string]struct {
		addr          string
		useReflection bool
		useTLS        bool
		cacert        string
		cert          string
		certKey       string

		err error
	}{
		"cert is missing":                         {useTLS: true, cert: "foo", cacert: "bar", err: entity.ErrMutualAuthParamsAreNotEnough},
		"cacert is missing":                       {useTLS: true, cert: "foo", certKey: "bar", err: entity.ErrMutualAuthParamsAreNotEnough},
		"certKey is missing":                      {useTLS: true, cacert: "foo", cert: "foo", err: entity.ErrMutualAuthParamsAreNotEnough},
		"cert is missing, but useTLS is false":    {cert: "foo", cacert: "bar"},
		"cacert is missing, but useTLS is false":  {cert: "foo", certKey: "bar"},
		"certKey is missing, but useTLS is false": {cacert: "foo", cert: "foo"},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			_, err := NewClient(c.addr, c.useReflection, c.useTLS, c.cacert, c.cert, c.certKey)
			if c.err != nil {
				require.Error(t, err, "NewClient must return an error")
				assert.Equal(t, c.err, err)
				return
			} else {
				require.NoError(t, err, "NewClient must not return an error")
			}
		})
	}
}
