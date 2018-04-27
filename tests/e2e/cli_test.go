package e2e

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/ktr0731/evans/adapter/controller"
	"github.com/ktr0731/evans/meta"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/test/grpc_testing"
)

func newCLI(t *testing.T, ui controller.UI) *controller.CLI {
	return controller.NewCLI(meta.AppName, meta.Version.String(), ui)
}

func TestCLI(t *testing.T) {
	in := strings.NewReader(`{ "name": "maho", "message": "hiyajo" }`)

	controller.DefaultCLIReader = in
	defer func() {
		controller.DefaultCLIReader = os.Stdin
	}()

	runServer := func() chan struct{} {
		grpc_testing.Empty{}
	}

	t.Run("from stdin", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := controller.NewUI(in, out, os.Stderr)
		c := newCLI(t, ui)
		code := c.Run([]string{"--package", "helloworld", "--service", "Greeter", "--call", "SayHello", "testdata/helloworld.proto"})
		require.Equal(t, 0, code)
		pp.Println(out.String())
	})
}
