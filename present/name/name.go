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

// Format formats v into the list of names. v should be a struct type.
// Format tries to find "name" tag from the struct fields and format the first appeared field.
// The struct type is only allowed to have struct, slice or primitive type fields.
func (p *Presenter) Format(v any) (string, error) {
	return formatFromStruct(reflect.ValueOf(v))
}

func formatFromStruct(rv reflect.Value) (string, error) {
	rv = indirect(rv)
	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		return "", errors.New("v should be a struct type")
	}

	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		v := f.Tag.Get("name")
		if v == "" {
			if f.Type.Kind() == reflect.Struct {
				s, err := formatFromStruct(rv.Field(i))
				if err == nil {
					return s, nil
				}
			}
			continue
		}

		if f.Type.Kind() == reflect.Slice {
			slice := rv.Field(i)
			rows := make([]string, slice.Len())
			for i := 0; i < slice.Len(); i++ {
				v := slice.Index(i)
				if v.Type().Kind() != reflect.Struct {
					return "", errors.New("v should have a slice of a struct")
				}
				if v.NumField() == 0 {
					return "", errors.New("struct should have at least 1 field")
				}
				for j := 0; j < v.NumField(); j++ {
					if v.Type().Field(j).Tag.Get("name") == "" {
						continue
					}
					rows[i] = fmt.Sprint(v.Field(j))
					break
				}
			}
			return strings.Join(rows, "\n"), nil
		}
		return fmt.Sprint(rv.Field(i)), nil
	}
	return "", errors.New("invalid type")
}

func NewPresenter() *Presenter {
	return &Presenter{}
}
