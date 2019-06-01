package app

import (
	"testing"
)

func Test_stringToStringSliceValue(t *testing.T) {
	m := make(map[string][]string)
	v := newStringToStringValue(map[string][]string{
		"ogiso": []string{"setsuna"},
	}, &m)
	if err := v.Set("touma=kazusa,touma=youko"); err != nil {
		t.Fatalf("Set must not return an error, but got '%s'", err)
	}
	const expected = `["touma=kazusa,youko"]`
	actual := v.String()
	if expected != actual {
		t.Errorf("expected '%s', but got '%s'", expected, actual)
	}
}
