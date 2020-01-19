package idl_test

import (
	"testing"

	"github.com/ktr0731/evans/idl"
)

func TestFullyQualifiedServiceName(t *testing.T) {
	cases := map[string]struct {
		pkg, svc    string
		expected    string
		expectedErr error
	}{
		"normal":             {pkg: "foo", svc: "Bar", expected: "foo.Bar"},
		"only service":       {svc: "Bar", expected: "Bar"},
		"service unselected": {pkg: "foo", expectedErr: idl.ErrServiceUnselected},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			fqsn, err := idl.FullyQualifiedServiceName(c.pkg, c.svc)
			if c.expectedErr != nil {
				if err != c.expectedErr {
					t.Errorf("expected error '%s', but got '%s'", c.expectedErr, err)
				}
				return
			}
			if fqsn != c.expected {
				t.Errorf("expected %s, but got %s", c.expected, fqsn)
			}
		})
	}
}
