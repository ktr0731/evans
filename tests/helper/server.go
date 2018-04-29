package helper

import srv "github.com/ktr0731/evans/tests/helper/server"

import (
	"net"
	"testing"

	"github.com/ktr0731/evans/tests/helper/server/helloworld"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type server struct {
	t *testing.T
	s *grpc.Server
}

func NewServer(t *testing.T) *server {
	s := grpc.NewServer()
	helloworld.RegisterGreeterServer(s, srv.NewUnary())
	return &server{
		t: t,
		s: s,
	}
}

func (s *server) Start() *server {
	go func() {
		l, err := net.Listen("tcp", ":50051")
		require.NoError(s.t, err)
		err = s.s.Serve(l)
		require.NoError(s.t, err)
	}()
	return s
}

func (s *server) Stop() {
	s.s.GracefulStop()
}
