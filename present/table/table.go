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

	var (
		keys []string
		vals [][]string
		err  error
	)
	rt := rv.Type()
	switch rt.Kind() {
	case reflect.Slice:
		it := indirectType(rt.Elem())
		if it.Kind() != reflect.Struct {
			return "", errors.New("v should be a slice of a struct type")
		}
		keys, err = processStructKeys(it)
		if err != nil {
			return "", errors.Wrap(err, "failed to get struct keys of the passed slice")
		}
		for i := 0; i < rv.Len(); i++ {
			rows, err := processStructValues(indirect(rv.Index(i)))
			if err != nil {
				return "", errors.Wrap(err, "failed to get struct values")
			}
			vals = append(vals, rows)
		}
	case reflect.Struct:
		keys, err = processStructKeys(rt)
		if err != nil {
			return "", errors.Wrap(err, "failed to get struct keys")
		}
		rows, err := processStructValues(rv)
		if err != nil {
			return "", errors.Wrap(err, "failed to get struct values")
		}
		vals = append(vals, rows)
	default:
		return "", errors.New("v should be a struct or a slice of a struct")
	}

	var w bytes.Buffer
	table := tablewriter.NewWriter(&w)
	table.SetHeader(keys)
	table.AppendBulk(vals)
	table.Render()
	return w.String(), nil
}

func processStructKeys(rt reflect.Type) ([]string, error) {
	if rt.Kind() != reflect.Struct {
		return nil, errors.New("v should be a struct type")
	}
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
	return keys, nil
}

func processStructValues(rv reflect.Value) ([]string, error) {
	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		return nil, errors.New("v should be a struct type")
	}
	var vals []string
	for i := 0; i < rt.NumField(); i++ {
		f := rv.Field(i)
		vals = append(vals, fmt.Sprint(f.Interface()))
	}
	return vals, nil
}

func NewPresenter() *Presenter {
	return &Presenter{}
}
