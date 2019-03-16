package grpc

import (
	"path/filepath"
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
	certPath := func(s ...string) string {
		return filepath.Join(append([]string{"testdata", "cert"}, s...)...)
	}
	cases := map[string]struct {
		addr          string
		useReflection bool
		useTLS        bool
		cacert        string
		cert          string
		certKey       string

		hasErr bool
		err    error
	}{
		"certKey is missing":                      {useTLS: true, cert: "foo", err: entity.ErrMutualAuthParamsAreNotEnough},
		"cert is missing":                         {useTLS: true, certKey: "bar", err: entity.ErrMutualAuthParamsAreNotEnough},
		"certKey is missing, but useTLS is false": {cert: "foo"},
		"cert is missing, but useTLS is false":    {certKey: "foo"},
		"enable server TLS":                       {useTLS: true},
		"enable server TLS with a trusted CA":     {useTLS: true, cacert: certPath("rootCA.pem")},
		"enable mutual TLS":                       {useTLS: true, cert: certPath("localhost.pem"), certKey: certPath("localhost-key.pem")},
		"enable mutual TLS with a trusted CA":     {useTLS: true, cacert: certPath("rootCA.pem"), cert: certPath("localhost.pem"), certKey: certPath("localhost-key.pem")},
		"invalid cacert file path":                {useTLS: true, cacert: "fooCA.pem", hasErr: true},
		"invalid cert and key file path":          {useTLS: true, cert: "foo.pem", certKey: "foo-key.pem", hasErr: true},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			_, err := NewClient(c.addr, "", c.useReflection, c.useTLS, c.cacert, c.cert, c.certKey)
			if c.err != nil {
				require.Error(t, err, "NewClient must return an error")
				assert.Equal(t, c.err, err)
				return
			} else if c.hasErr {
				require.Error(t, err, "NewClient must return an error")
				return
			}
			require.NoError(t, err, "NewClient must not return an error")
		})
	}
}
