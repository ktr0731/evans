package app

import (
	"testing"
)

func Test_stringToStringSliceValue(t *testing.T) {
	m := make(map[string][]string)
	v := newStringToStringValue(map[string][]string{
		"ogiso": []string{"setsuna"},
	}, &m)
	v.Set("touma=kazusa,touma=youko")
	const expected = `["touma=kazusa,youko"]`
	actual := v.String()
	if expected != actual {
		t.Errorf("expected '%s', but got '%s'", expected, actual)
	}
}
