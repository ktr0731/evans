package name

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Presenter is a presenter that formats v into the list of names.
type Presenter struct{}

func indirect(rv reflect.Value) reflect.Value {
	if rv.Type().Kind() != reflect.Ptr {
		return rv
	}
	return indirect(reflect.Indirect(rv))
}

func findSlice(rv reflect.Value) (reflect.Value, bool) {
	for i := 0; i < rv.NumField(); i++ {
		sf := rv.Field(i)
		if sf.Kind() == reflect.Slice {
			return sf, true
		}
	}
	return rv, false
}

// Format formats v into the list of names. v should be a struct type that have a struct slice.
func (p *Presenter) Format(v interface{}, indent string) (string, error) {
	rv := indirect(reflect.ValueOf(v))
	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		return "", errors.New("v should be a struct type")
	}

	slice, ok := findSlice(rv)
	if !ok {
		return "", errors.New("the struct should have a slice field")
	}

	rows := make([]string, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		v := slice.Index(i)
		if v.Type().Kind() != reflect.Struct {
			return "", errors.New("v should have a slice of a struct")
		}
		if v.NumField() == 0 {
			return "", errors.New("struct should have at least 1 field")
		}
		rows[i] = fmt.Sprint(v.Field(0))
	}
	return strings.Join(rows, "\n"), nil
}

func NewPresenter() *Presenter {
	return &Presenter{}
}
