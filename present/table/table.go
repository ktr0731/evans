// Package table provides a table like formatting.
package table

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// Presenter is a presenter that formats v with table layout.
type Presenter struct{}

func indirect(rv reflect.Value) reflect.Value {
	if rv.Type().Kind() != reflect.Ptr {
		return rv
	}
	return indirect(reflect.Indirect(rv))
}

func indirectType(rt reflect.Type) reflect.Type {
	if rt.Kind() != reflect.Ptr {
		return rt
	}
	return indirectType(rt.Elem())
}

// Format formats v with table layout. v should be a struct type that has a slice field. Format extract the slice field
// and display them. See test cases for example.
func (p *Presenter) Format(v interface{}) (string, error) {
	rv := indirect(reflect.ValueOf(v))
	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		return "", errors.New("v should be a struct type")
	}

	slice, ok := findSlice(rv)
	if !ok {
		return "", errors.New("the struct should have a slice field")
	}

	keys := processStructKeys(slice.Type().Elem())
	rows := make([][]string, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		rows[i] = processStructValues(slice.Index(i))
	}

	var w bytes.Buffer
	table := tablewriter.NewWriter(&w)
	table.SetHeader(keys)
	table.AppendBulk(rows)
	table.Render()
	return w.String(), nil
}

func findSlice(rv reflect.Value) (_ reflect.Value, ok bool) {
	for i := 0; i < rv.NumField(); i++ {
		sf := rv.Field(i)
		if sf.Kind() == reflect.Slice {
			return sf, true
		}
	}
	return rv, false
}

func processStructKeys(rt reflect.Type) []string {
	rt = indirectType(rt)
	keys := make([]string, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		key := sf.Tag.Get("table")
		if key == "-" {
			continue
		}
		if key == "" {
			key = strings.ToLower(sf.Name)
		}
		keys = append(keys, key)
	}
	return keys
}

func processStructValues(rv reflect.Value) []string {
	rv = indirect(rv)
	row := make([]string, rv.NumField())
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		row[i] = fmt.Sprint(f.Interface())
	}
	return row
}

func NewPresenter() *Presenter {
	return &Presenter{}
}
