package grpc

import (
	"context"
	"io"

	"github.com/ktr0731/evans/grpc/grpcreflection"
	"github.com/ktr0731/grpc-web-go-client/grpcweb"
	"github.com/pkg/errors"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type webClient struct {
	conn    *grpcweb.ClientConn
	headers Headers

	grpcreflection.Client
}

func NewWebClient(addr string, useReflection, useTLS bool, cacert, cert, certKey string) Client {
	conn, err := grpcweb.DialContext(addr)
	if err != nil {
		panic(err)
	}
	client := &webClient{
		conn:    conn,
		headers: Headers{},
	}

	if useReflection {
		client.Client = grpcreflection.NewWebClient(conn)
	}

	return client
}

func (c *webClient) Invoke(ctx context.Context, fqrn string, req, res interface{}, opts ...gogrpc.CallOption) error {
	endpoint, err := fqrnToEndpoint(fqrn)
	if err != nil {
		return errors.Wrap(err, "grpc-web: failed to convert FQRN to endpoint")
	}

	loggingRequest(req)

	err = c.conn.Invoke(ctx, endpoint, req, res)
	// TODO: Handle gRPC-Web misconfiguration errors for more helpful message.
	if err != nil {
		return errors.Wrap(err, "grpc-web: failed to send a request")
	}
	return nil
}

type webClientStream struct {
	ctx    context.Context
	stream grpcweb.ClientStream
}

func (s *webClientStream) Header() (metadata.MD, error) {
	return nil, errors.New("not supported yet")
}

func (s *webClientStream) Trailer() metadata.MD {
	panic("not supported yet")
}

func (s *webClientStream) Send(req interface{}) error {
	loggingRequest(req)
	if err := s.stream.Send(s.ctx, req); err != nil {
		return errors.Wrap(err, "failed to send a request")
	}
	return nil
}

func (s *webClientStream) CloseAndReceive(res interface{}) error {
	err := s.stream.CloseAndReceive(s.ctx, res)
	if err != nil {
		return errors.Wrap(err, "failed to send CloseAndReceive")
	}
	return nil
}

func (c *webClient) NewClientStream(ctx context.Context, streamDesc *gogrpc.StreamDesc, fqrn string) (ClientStream, error) {
	endpoint, err := fqrnToEndpoint(fqrn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert FQRN to endpoint")
	}

	stream, err := c.conn.NewClientStream(streamDesc, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a new client stream")
	}
	return &webClientStream{
		ctx:    ctx,
		stream: stream,
	}, nil
}

type webServerStream struct {
	ctx    context.Context
	stream grpcweb.ServerStream
}

func (s *webServerStream) Header() (metadata.MD, error) {
	return nil, errors.New("not supported yet")
}

func (s *webServerStream) Trailer() metadata.MD {
	panic("not supported yet")
}

func (s *webServerStream) Send(req interface{}) (err error) {
	loggingRequest(req)
	if err := s.stream.Send(s.ctx, req); err != nil {
		return errors.Wrap(err, "failed to send a request")
	}
	return nil
}

func (s *webServerStream) Receive(res interface{}) error {
	if s.stream == nil {
		return errors.New("Receive must be call after Send method")
	}
	err := s.stream.Receive(s.ctx, res)
	if err != nil {
		return err
	}
	return nil
}

func (c *webClient) NewServerStream(ctx context.Context, streamDesc *gogrpc.StreamDesc, fqrn string) (ServerStream, error) {
	endpoint, err := fqrnToEndpoint(fqrn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert FQRN to endpoint")
	}

	stream, err := c.conn.NewServerStream(streamDesc, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a new server stream")
	}

	return &webServerStream{ctx: ctx, stream: stream}, nil
}

type webBidiStream struct {
	ctx    context.Context
	stream grpcweb.BidiStream
}

func (s *webBidiStream) Header() (metadata.MD, error) {
	return nil, errors.New("not supported yet")
}

func (s *webBidiStream) Trailer() metadata.MD {
	panic("not supported yet")
}

func (s *webBidiStream) Send(req interface{}) error {
	loggingRequest(req)
	if err := s.stream.Send(s.ctx, req); err != nil {
		return errors.Wrap(err, "failed to send a request")
	}
	return nil
}

func (s *webBidiStream) Receive(res interface{}) error {
	err := s.stream.Receive(s.ctx, res)
	if errors.Cause(err) == io.EOF {
		return io.EOF
	}
	if err != nil {
		return errors.Wrap(err, "failed to receive a response")
	}
	return nil
}

func (s *webBidiStream) CloseSend() error {
	if err := s.stream.CloseSend(); err != nil {
		return errors.Wrap(err, "failed to close the send stream")
	}
	return nil
}

func (c *webClient) NewBidiStream(ctx context.Context, streamDesc *gogrpc.StreamDesc, fqrn string) (BidiStream, error) {
	endpoint, err := fqrnToEndpoint(fqrn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert FQRN to endpoint")
	}

	stream, err := c.conn.NewBidiStream(streamDesc, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a new bidi stream")
	}

	return &webBidiStream{
		ctx:    ctx,
		stream: stream,
	}, nil
}

func (c *webClient) Close(ctx context.Context) error {
	if c.Client != nil {
		c.Client.Reset()
	}
	return nil
}

func (c *webClient) Headers() Headers {
	return c.headers
}
