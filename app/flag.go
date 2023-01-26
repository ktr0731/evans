package app

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/ktr0731/go-multierror"
	"github.com/pkg/errors"
)

//nolint:maligned
// flags defines available command line flags.
type flags struct {
	mode struct {
		repl bool
		cli  bool
	}

	cli struct {
		call string
		file string
	}

	repl struct {
		silent bool
	}

	common struct {
		pkg        string
		service    string
		path       []string
		proto      []string
		host       string
		port       string
		header     map[string][]string
		web        bool
		reflection bool
		tls        bool
		cacert     string
		cert       string
		certKey    string
		serverName string
		insecure   bool
	}

	meta struct {
		edit       bool
		editGlobal bool
		verbose    bool
		version    bool
		help       bool
	}
}

// validate defines invalid conditions and validates whether f has invalid conditions.
func (f *flags) validate() error {
	var result error
	invalidCases := []struct {
		name string
		cond bool
	}{
		{"cannot specify both of --cli and --repl", f.mode.cli && f.mode.repl},
	}
	for _, c := range invalidCases {
		if c.cond {
			result = multierror.Append(result, errors.New(c.name))
		}
	}
	return result
}

// -- stringToString Value.
type stringToStringSliceValue struct {
	value   *map[string][]string
	changed bool
}

func newStringToStringValue(val map[string][]string, p *map[string][]string) *stringToStringSliceValue {
	ssv := new(stringToStringSliceValue)
	ssv.value = p
	*ssv.value = val
	return ssv
}

// Format: a=1,b=2.
func (s *stringToStringSliceValue) Set(val string) error {
	var ss []string
	n := strings.Count(val, "=")
	switch n {
	case 0:
		return errors.Errorf("%s must be formatted as key=value", val)
	case 1:
		ss = append(ss, strings.Trim(val, `"`))
	default:
		r := csv.NewReader(strings.NewReader(val))
		var err error
		ss, err = r.Read()
		if err != nil {
			return err
		}
	}

	out := make(map[string][]string, len(ss))
	for _, pair := range ss {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("%s must be formatted as key=value", pair)
		}
		out[kv[0]] = append(out[kv[0]], kv[1])
	}
	if !s.changed {
		*s.value = out
	} else {
		for k, v := range out {
			(*s.value)[k] = v
		}
	}
	s.changed = true
	return nil
}

func (s *stringToStringSliceValue) Type() string {
	return "slice of strings"
}

func (s *stringToStringSliceValue) String() string {
	records := make([]string, 0, len(*s.value)>>1)
	for k, v := range *s.value {
		records = append(records, k+"="+strings.Join(v, ","))
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(records); err != nil {
		panic(err)
	}
	w.Flush()
	return "[" + strings.TrimSpace(buf.String()) + "]"
}
