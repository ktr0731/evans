package app

import (
	"testing"
)

func Test_stringToStringSliceValue(t *testing.T) {
	cases := []struct {
		in       string
		expected string
		hasErr   bool
	}{
		{"touma=kazusa,touma=youko", `["touma=kazusa,youko"]`, false},
		{"sawamura='spencer=eriri'", `[sawamura='spencer=eriri']`, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.in, func(t *testing.T) {
			m := make(map[string][]string)
			v := newStringToStringValue(map[string][]string{
				"ogiso": []string{"setsuna"},
			}, &m)
			if err := v.Set(c.in); err != nil {
				t.Fatalf("Set must not return an error, but got '%s'", err)
			}
			actual := v.String()
			if c.expected != actual {
				t.Errorf("expected '%s', but got '%s'", c.expected, actual)
			}
		})
	}
}
