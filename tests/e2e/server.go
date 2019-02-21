// +build e2e

package e2e

import (
	"context"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	srv "github.com/ktr0731/evans/tests/e2e/server"
	"github.com/ktr0731/evans/tests/e2e/server/helloworld"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type server struct {
	t  *testing.T
	s  *grpc.Server
	ws *http.Server

	errMu sync.Mutex
	err   error

	port string
}

func newServer(t *testing.T, useReflection bool, useTLS bool) *server {
	var opts []grpc.ServerOption
	if useTLS {
		creds, err := credentials.NewServerTLSFromFile(filepath.Join("testdata", "cert", "localhost.pem"), filepath.Join("testdata", "cert", "localhost-key.pem"))
		require.NoError(t, err)
		opts = append(opts, grpc.Creds(creds))
	}
	s := grpc.NewServer(opts...)
	helloworld.RegisterGreeterServer(s, srv.NewUnary())
	if useReflection {
		reflection.Register(s)
	}
	port, err := freeport.GetFreePort()
	require.NoError(t, err, "failed to get a free port")
	return &server{
		t:    t,
		s:    s,
		port: strconv.Itoa(port),
	}
}

func (s *server) start(web bool) *server {
	addr := ":" + s.port
	if web {
		ws := grpcweb.WrapServer(
			s.s,
			grpcweb.WithWebsockets(true),
			grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool { return true }),
		)
		mux := http.NewServeMux()
		mux.Handle("/", ws)
		s.ws = &http.Server{
			Addr:    addr,
			Handler: mux,
		}

		go func() {
			err := s.ws.ListenAndServe()
			if err == http.ErrServerClosed {
				return
			}
			if err != nil {
				s.reportError(err)
				s.t.Fail()
				return
			}
		}()

		return s
	}

	l, err := net.Listen("tcp", addr)
	require.NoError(s.t, err)
	go func() {
		err = s.s.Serve(l)
		if err != nil && err != grpc.ErrServerStopped {
			s.reportError(err)
			s.t.Fail()
			return
		}
	}()
	return s
}

func (s *server) stop() {
	if s.gRPCWebEnabled() {
		s.ws.Shutdown(context.Background())
		return
	}
	s.s.GracefulStop()

	s.errMu.Lock()
	defer s.errMu.Unlock()
	if s.t.Failed() {
		s.t.Error(s.err)
	}
}

func (s *server) gRPCWebEnabled() bool {
	return s.ws != nil
}

func (s *server) reportError(err error) {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	s.err = multierror.Append(s.err, err)
}
