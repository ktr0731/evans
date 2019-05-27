package app

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/go-multierror"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

var usageFormat = `
Usage: %s [--help] [--version] [options ...] [PROTO [PROTO ...]]

Positional arguments:
        PROTO                          .proto files

Options:
%s
`

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

func (a *App) parseFlags(args []string) (*flags, error) {
	f := pflag.NewFlagSet("main", pflag.ContinueOnError)
	f.SortFlags = false
	f.SetOutput(a.cui.Writer)

	var flags flags

	f.BoolVar(&flags.mode.repl, "repl", false, "launch Evans as REPL mode")
	f.BoolVar(&flags.mode.cli, "cli", false, "start as CLI mode")

	f.StringVar(&flags.cli.call, "call", "", "call specified RPC by CLI mode")
	f.StringVarP(&flags.cli.file, "file", "f", "", "a script file that will be executed by (used only CLI mode)")

	f.BoolVarP(&flags.repl.silent, "silent", "s", false, "hide redundant output")

	f.StringVar(&flags.common.pkg, "package", "", "default package")
	f.StringVar(&flags.common.service, "service", "", "default service")
	f.StringSliceVar(&flags.common.path, "path", nil, "proto file paths")
	f.StringVar(&flags.common.host, "host", "", "gRPC server host")
	f.StringVarP(&flags.common.port, "port", "p", "50051", "gRPC server port")
	f.Var(
		newStringToStringValue(nil, &flags.common.header),
		"header", "default headers that set to each requests (example: foo=bar)")
	f.BoolVar(&flags.common.web, "web", false, "use gRPC-Web protocol")
	f.BoolVarP(&flags.common.reflection, "reflection", "r", false, "use gRPC reflection")
	f.BoolVarP(&flags.common.tls, "tls", "t", false, "use a secure TLS connection")
	f.StringVar(&flags.common.cacert, "cacert", "", "the CA certificate file for verifying the server")
	f.StringVar(
		&flags.common.cert,
		"cert", "", "the certificate file for mutual TLS auth. it must be provided with --certkey.")
	f.StringVar(
		&flags.common.certKey,
		"certkey", "", "the private key file for mutual TLS auth. it must be provided with --cert.")
	f.StringVar(
		&flags.common.serverName,
		"servername", "", "override the server name used to verify the hostname (ignored if --tls is disabled)")

	f.BoolVarP(&flags.meta.edit, "edit", "e", false, "edit the project config file by using $EDITOR")
	f.BoolVar(&flags.meta.editGlobal, "edit-global", false, "edit the global config file by using $EDITOR")
	f.BoolVar(&flags.meta.verbose, "verbose", false, "verbose output")
	f.BoolVarP(&flags.meta.version, "version", "v", false, "display version and exit")
	f.BoolVarP(&flags.meta.help, "help", "h", false, "display help text and exit")

	f.Usage = func() {
		a.printVersion()
		var buf bytes.Buffer
		w := tabwriter.NewWriter(&buf, 0, 8, 8, ' ', tabwriter.TabIndent)
		f.VisitAll(func(f *pflag.Flag) {
			cmd := "--" + f.Name
			if f.Shorthand != "" {
				cmd += ", -" + f.Shorthand
			}
			name, _ := pflag.UnquoteUsage(f)
			if name != "" {
				cmd += " " + name
			}
			usage := f.Usage
			if f.DefValue != "" {
				usage += fmt.Sprintf(` (default "%s")`, f.DefValue)
			}
			fmt.Fprintf(w, "        %s\t%s\n", cmd, usage)
		})
		w.Flush()
		fmt.Fprintf(a.cui.Writer, usageFormat, meta.AppName, buf.String())
	}

	// ignore error because flag set mode is ExitOnError
	err := f.Parse(args)
	if err != nil {
		return nil, err
	}

	a.flagSet = f

	return &flags, nil
}

// -- stringToString Value
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

// Format: a=1,b=2
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
		kv := strings.Split(pair, "=")
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
