package helper

import srv "github.com/ktr0731/evans/tests/helper/server"

import (
	"net"
	"testing"

	"github.com/ktr0731/evans/tests/helper/server/helloworld"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type Server struct {
	t *testing.T
	s *grpc.Server
}

func NewServer(t *testing.T) *Server {
	s := grpc.NewServer()
	helloworld.RegisterGreeterServer(s, srv.NewUnary())
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
