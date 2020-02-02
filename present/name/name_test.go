package name_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/present/name"
)

func TestPresenter(t *testing.T) {
	cases := map[string]struct {
		v        interface{}
		expected string
		hasErr   bool
	}{
		"not a struct": {
			v:      100,
			hasErr: true,
		},
		"doesn't have a slice": {
			v:      struct{}{},
			hasErr: true,
		},
		"doesn't have a slice of a struct": {
			v:      struct{ V []int }{[]int{1}},
			hasErr: true,
		},
		"the slice type has no fields": {
			v: struct {
				V []struct{}
			}{
				V: []struct{}{struct{}{}},
			},
			hasErr: true,
		},
		"normal": {
			v: struct {
				V []struct{ int }
			}{
				V: []struct{ int }{struct{ int }{100}, struct{ int }{200}},
			},
			expected: "100\n200",
		},
	}
	for tname, c := range cases {
		t.Run(tname, func(t *testing.T) {
			p := name.NewPresenter()
			actual, err := p.Format(c.v, "")
			if c.hasErr {
				if err == nil {
					t.Errorf("should return an error, but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("should not return an error, but got '%s'", err)
			}
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("(-want, +got)\n%s", diff)
			}
		})
	}
}
