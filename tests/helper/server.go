package helper

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	srv "github.com/ktr0731/evans/tests/helper/server"
	"github.com/ktr0731/evans/tests/helper/server/helloworld"
	multierror "github.com/ktr0731/go-multierror"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	t  *testing.T
	s  *grpc.Server
	ws *http.Server

	errMu sync.Mutex
	err   error
}

func NewServer(t *testing.T, enableReflection bool) *Server {
	s := grpc.NewServer()
	helloworld.RegisterGreeterServer(s, srv.NewUnary())
	if enableReflection {
		reflection.Register(s)
	}
	return &Server{
		t: t,
		s: s,
	}
}

func (s *Server) Start(web bool) *Server {
	if web {
		ws := grpcweb.WrapServer(
			s.s,
			grpcweb.WithWebsockets(true),
			grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool { return true }),
		)
		mux := http.NewServeMux()
		mux.Handle("/", ws)
		s.ws = &http.Server{
			Addr:    ":50051",
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

	l, err := net.Listen("tcp", ":50051")
	require.NoError(s.t, err)
	go func() {
		err = s.s.Serve(l)
		if err != nil {
			s.reportError(err)
			s.t.Fail()
			return
		}
	}()
	return s
}

func (s *Server) Stop() {
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

func (s *Server) gRPCWebEnabled() bool {
	return s.ws != nil
}

func (s *Server) reportError(err error) {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	s.err = multierror.Append(s.err, err)
}
