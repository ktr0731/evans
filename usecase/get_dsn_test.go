package usecase

import "testing"

func TestGetDomainSourceName(t *testing.T) {
	cases := map[string]struct {
		pkg, svc string
		expected string
	}{
		"both":         {pkg: "foo", svc: "Bar", expected: "foo.Bar"},
		"only package": {pkg: "foo", expected: "foo"},
		"only service": {svc: "Bar", expected: "Bar"},
		"nothing":      {expected: ""},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			d := &dependencyManager{
				state: state{
					selectedPackage: c.pkg,
					selectedService: c.svc,
				},
			}
			dsn := d.GetDomainSourceName()
			if dsn != c.expected {
				t.Errorf("expected %s, but got %s", c.expected, dsn)
			}
		})
	}
}
