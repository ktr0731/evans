package fill_test

import (
	"strings"
	"testing"

	"github.com/ktr0731/evans/fill"
)

func TestSilentFiller(t *testing.T) {
	cases := map[string]struct {
		in     string
		hasErr bool
	}{
		"normal":       {in: `{"foo": "bar"}`},
		"invalid JSON": {in: `foo`, hasErr: true},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			f := fill.NewSilentFiller(strings.NewReader(c.in))
			var i interface{}
			err := f.Fill(&i)
			if c.hasErr {
				if err == nil {
					t.Errorf("Fill must return an error, but got nil")
				}
			} else if err != nil {
				t.Errorf("Fill must not return an error, but got an error: '%s'", err)
			}
		})
	}
}
