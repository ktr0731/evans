// +build e2e

package e2e

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ktr0731/grpc-test/server"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(
		m,
		// TODO: invest these leaks
		goleak.IgnoreTopFunction("github.com/ktr0731/evans/vendor/google.golang.org/grpc"),
		goleak.IgnoreTopFunction("github.com/ktr0731/evans/vendor/google.golang.org/grpc.(*ccBalancerWrapper).watcher"),
		goleak.IgnoreTopFunction("github.com/ktr0731/evans/vendor/google.golang.org/grpc.(*ccResolverWrapper).watcher"),
		goleak.IgnoreTopFunction("github.com/ktr0731/evans/vendor/google.golang.org/grpc.(*addrConn).createTransport"),

		// ref. repl.(*executor).execute comments
		goleak.IgnoreTopFunction("time.Sleep"),
		goleak.IgnoreTopFunction("runtime.goparkunlock"),
	)
}

func newServer(t *testing.T, useReflection, useTLS, useWeb bool) (*server.Server, string) {
	port, err := freeport.GetFreePort()
	require.NoError(t, err, "failed to get a free port for gRPC test server")

	addr := fmt.Sprintf(":%d", port)
	opts := []server.Option{server.WithAddr(addr)}
	if useReflection {
		opts = append(opts, server.WithReflection())
	}
	if useTLS {
		opts = append(opts, server.WithTLS())
	}
	if useWeb {
		opts = append(opts, server.WithProtocol(server.ProtocolImprobableGRPCWeb))
	}

	return server.New(opts...), strconv.Itoa(port)
}
