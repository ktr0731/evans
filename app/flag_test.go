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
		{in: "touma=kazusa,touma=youko", expected: `["touma=kazusa,youko"]`, hasErr: false},
		{in: "sawamura='spencer=eriri'", expected: `[sawamura='spencer=eriri']`, hasErr: false},
		{in: "sawamura=spencer=eriri", expected: `[sawamura=spencer=eriri]`, hasErr: false},
		{in: "megumi=kato", expected: `[megumi=kato]`, hasErr: false},
		{in: "yuki=asuna,alice", expected: `["yuki=asuna,alice"]`, hasErr: false},
		{in: "alice", hasErr: true},
	}
	for _, c := range cases {
		c := c
		t.Run(c.in, func(t *testing.T) {
			m := make(map[string][]string)
			v := newStringToStringValue(map[string][]string{
				"ogiso": []string{"setsuna"},
			}, &m)
			err := v.Set(c.in)
			if c.hasErr {
				if err == nil {
					t.Fatalf("Set must return an error, but got nil")
				}
				return
			} else if err != nil {
				t.Fatalf("Set must not return an error, but got '%s'", err)
			}
			actual := v.String()
			if c.expected != actual {
				t.Errorf("expected '%s', but got '%s'", c.expected, actual)
			}
		})
	}
}
