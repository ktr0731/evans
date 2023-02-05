package repl

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/grpc"
	"github.com/ktr0731/evans/prompt"
	"github.com/ktr0731/evans/usecase"
)

func TestREPL_helpText(t *testing.T) {
	dummyCfg := &config.Config{REPL: &config.REPL{}, Server: &config.Server{Host: "127.0.0.1", Port: "50051"}}

	usecase.Clear()

	r, err := New(dummyCfg, prompt.New(), nil, "", "")
	if err != nil {
		t.Fatalf("New must not return an erorr, but got '%s'", err)
	}

	actual := r.helpText()
	if diff := cmp.Diff(expectedHelpText, actual); diff != "" {
		t.Errorf("diff found:\n%s", diff)
	}
}

func TestREPL_printSplash(t *testing.T) {
	dummyCfg := &config.Config{REPL: &config.REPL{}, Server: &config.Server{Host: "127.0.0.1", Port: "50051"}}

	usecase.Clear()

	w := new(bytes.Buffer)
	ui := cui.New(cui.Writer(w))

	r, err := New(dummyCfg, prompt.New(), ui, "", "")
	if err != nil {
		t.Fatalf("New must not return an erorr, but got '%s'", err)
	}

	r.printSplash("")
	if diff := cmp.Diff(w.String(), defaultSplashText+"\n"); diff != "" {
		t.Errorf("unexpected default splash: %s", diff)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home dir: %s", err)
	}

	f, err := os.CreateTemp(home, "")
	if err != nil {
		t.Fatalf("failed to create a temp file: %s", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if _, err := io.WriteString(f, "hi."); err != nil {
		t.Fatalf("WriteString must not return an error, but got %s", err)
	}

	w.Reset()
	r.printSplash(f.Name())
	if w.String() != "hi.\n" {
		t.Errorf("unexpected splash: %s", w.String())
	}

	if runtime.GOOS != "windows" {
		w.Reset()
		// Replace with ~ for testing ~/ replacing logic.
		path := strings.Replace(f.Name(), home, "~", 1)
		r.printSplash(path)
		if w.String() != "hi.\n" {
			t.Errorf("unexpected splash: %s", w.String())
		}
	}
}

func TestREPL_makePrefix(t *testing.T) {
	cases := map[string]struct {
		pkgName string
		svcName string
		RPCsErr error

		hasErr   bool
		expected string
	}{
		"package and service unselected": {expected: "127.0.0.1:50051> "},
		"package selected":               {pkgName: "api", expected: "api@127.0.0.1:50051> "},
		"package and service selected": {
			pkgName:  "api",
			svcName:  "Example",
			expected: "api.Example@127.0.0.1:50051> ",
		},
	}

	for name, c := range cases {
		c := c
		dummyCfg := &config.Config{
			REPL:   &config.REPL{},
			Server: &config.Server{Host: "127.0.0.1", Port: "50051"},
		}
		dummySpec := &SpecMock{
			ServiceNamesFunc: func() []string {
				return []string{"api.Example"}
			},
			RPCsFunc: func(svcName string) ([]*grpc.RPC, error) {
				return nil, c.RPCsErr
			},
		}
		t.Run(name, func(t *testing.T) {
			usecase.Inject(usecase.Dependencies{Spec: dummySpec})

			r, err := New(dummyCfg, prompt.New(), nil, c.pkgName, c.svcName)
			if c.hasErr {
				if err == nil {
					t.Errorf("New must return an error, but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("New must not return an erorr, but got '%s'", err)
			}

			actual := r.makePrefix()
			if c.expected != actual {
				t.Errorf("expected prefix '%s', but got '%s'", c.expected, actual)
			}
		})
	}
}

var expectedHelpText = `
Available commands:
  call       call a RPC
  desc       describe the structure of selected message
  exit       exit current REPL
  header     set/unset headers to each request. if header value is empty, the header is removed.
  package    set a package as the currently selected package
  service    set the service as the current selected service
  show       show package, service or RPC names

Show more details:
  <command> --help`
