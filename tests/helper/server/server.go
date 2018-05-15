package server

import (
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/ktr0731/evans/tests/helper/server/helloworld"
	stream "github.com/ktr0731/evans/tests/helper/server/stream"
	context "golang.org/x/net/context"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// UnaryServer is an implementation of Greeter service
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

// StreamingServer is an implementation of Greeter service
// in tests/e2e/testdata/stream.proto
type StreamingServer struct{}

func NewStreaming() stream.GreeterServer {
	return &StreamingServer{}
}

func (s *StreamingServer) SayHelloClientStreaming(stm stream.Greeter_SayHelloClientStreamingServer) error {
	var t int
	var name string
	for {
		req, err := stm.Recv()
		if err == io.EOF {
			return stm.SendAndClose(&stream.HelloResponse{
				Message: fmt.Sprintf(`%s, you greet %d times.`, name, t),
			})
		}
		if err != nil {
			return err
		}
		name = req.GetName()
		t++
	}
}

func (s *StreamingServer) SayHelloServerStreaming(req *stream.HelloRequest, stm stream.Greeter_SayHelloServerStreamingServer) error {
	n := rand.Intn(10)
	for i := 0; i < n; i++ {
		err := stm.Send(&stream.HelloResponse{
			Message: fmt.Sprintf(`hello %s, I greet %d times.`, req.GetName(), i),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
