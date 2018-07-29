package helper

import (
	"net"
	"testing"

	srv "github.com/ktr0731/evans/tests/helper/server"
	"github.com/ktr0731/evans/tests/helper/server/helloworld"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	t *testing.T
	s *grpc.Server
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

func (s *Server) Start() *Server {
	l, err := net.Listen("tcp", ":50051")
	require.NoError(s.t, err)
	go func() {
		err = s.s.Serve(l)
		require.NoError(s.t, err)
	}()
	return s
}

func (s *Server) Stop() {
	s.s.GracefulStop()
}
