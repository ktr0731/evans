package server

import (
	"fmt"

	"github.com/ktr0731/evans/tests/helper/server/helloworld"
	context "golang.org/x/net/context"
)

// unaryServer is an implementation of Greeter service
// in tests/e2e/testdata/helloworld.proto
type UnaryServer struct{}

func NewUnary() helloworld.GreeterServer {
	return &UnaryServer{}
}

func (s *UnaryServer) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloResponse, error) {
	return &helloworld.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}
