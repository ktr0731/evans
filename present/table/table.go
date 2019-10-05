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

// TODO: remove indent.
func (p *Presenter) Format(v interface{}, indent string) (string, error) {
	rv := indirect(reflect.ValueOf(v))
	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		return "", errors.New("v should be a struct type")
	}

	keys := processStructKeys(rt)
	vals := processStructValues(rv)

	var w bytes.Buffer
	table := tablewriter.NewWriter(&w)
	table.SetHeader(keys)
	table.AppendBulk(vals)
	table.Render()
	return w.String(), nil
}

func processStructKeys(rt reflect.Type) []string {
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

func processStructValues(rv reflect.Value) [][]string {
	maxLen := 1
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		if f.Kind() == reflect.Slice && f.Len() > maxLen {
			maxLen = f.Len()
		}
	}

	var vals [][]string
	for i := 0; i < maxLen; i++ {
		row := make([]string, rv.NumField())
		for j := 0; j < rv.NumField(); j++ {
			f := rv.Field(j)
			if f.Kind() == reflect.Slice {
				if f.Len() > i {
					row[j] = fmt.Sprint(f.Index(i).Interface())
				}
			} else {
				row[j] = fmt.Sprint(f.Interface())
			}
		}
		vals = append(vals, row)
	}
	return vals
}

func NewPresenter() *Presenter {
	return &Presenter{}
}
