package grpc

import (
	"path/filepath"
	"testing"
)

func Test_fqrnToEndpoint(t *testing.T) {
	fqrn := "helloworld.Greeter.SayHello"

	endpoint, err := fqrnToEndpoint(fqrn)
	if err != nil {
		t.Fatalf("must not return an error, but got '%s'", err)
	}
	if expected := "/helloworld.Greeter/SayHello"; expected != endpoint {
		t.Fatalf("expected endpoint: %s, but got '%s'", expected, endpoint)
	}
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
		"certKey is missing":                      {useTLS: true, cert: "foo", err: ErrMutualAuthParamsAreNotEnough},
		"cert is missing":                         {useTLS: true, certKey: "bar", err: ErrMutualAuthParamsAreNotEnough},
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
			_, err := NewClient(c.addr, "", c.useReflection, c.useTLS, c.cacert, c.cert, c.certKey, nil)
			if c.err != nil {
				if err == nil {
					t.Fatalf("NewClient must return an error, but got nil")
				}
				if c.err != err {
					t.Errorf("expected: '%s', but got '%s'", c.err, err)
				}

				return
			} else if c.hasErr {
				if err == nil {
					t.Fatalf("NewClient must return an error, but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("NewClient must not return an error, but got '%s'", err)
			}
		})
	}
}
