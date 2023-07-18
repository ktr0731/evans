package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"crypto/tls"
	"crypto/x509"

	"github.com/hashicorp/go-multierror"
	"github.com/ktr0731/evans/grpc/grpcreflection"
	"github.com/ktr0731/evans/logger"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var ErrMutualAuthParamsAreNotEnough = errors.New("cert and certkey are required to authenticate mutually")

// RPC represents a RPC which belongs to a gRPC service.
type RPC struct {
	Name               string
	FullyQualifiedName string
	RequestType        *Type
	ResponseType       *Type
	IsServerStreaming  bool
	IsClientStreaming  bool
}

// Type is a type for representing requests/responses.
type Type struct {
	// Name is the name of Type.
	Name string

	// FullyQualifiedName is the name that contains the package name this Type belongs.
	FullyQualifiedName string

	// New instantiates a new instance of Type.  It is used for decode requests and responses.
	New func() interface{}
}

// Client represents the gRPC client.
type Client interface {
	// Invoke invokes a request req to the gRPC server. Then, Invoke decodes the response to res.
	Invoke(ctx context.Context, fqrn string, req, res interface{}) (header, trailer metadata.MD, _ error)

	// NewClientStream creates a new client stream.
	NewClientStream(ctx context.Context, streamDesc *grpc.StreamDesc, fqrn string) (ClientStream, error)

	// NewServerStream creates a new server stream.
	NewServerStream(ctx context.Context, streamDesc *grpc.StreamDesc, fqrn string) (ServerStream, error)

	// NewBidiStream creates a new bidirectional stream.
	NewBidiStream(ctx context.Context, streamDesc *grpc.StreamDesc, fqrn string) (BidiStream, error)

	// Close closes all connections the client has.
	Close(ctx context.Context) error

	// Header returns all request headers (metadata) Client has.
	Header() Headers

	grpcreflection.Client
}

type ClientStream interface {
	// Header returns the response header.
	Header() (metadata.MD, error)
	// Trailer returns the response trailer.
	Trailer() metadata.MD
	Send(req interface{}) error
	CloseAndReceive(res interface{}) error
}

type ServerStream interface {
	// Header returns the response header.
	Header() (metadata.MD, error)
	// Trailer returns the response trailer.
	Trailer() metadata.MD
	Send(req interface{}) error
	Receive(res interface{}) error
}

type BidiStream interface {
	// Header returns the response header.
	Header() (metadata.MD, error)
	// Trailer returns the response trailer.
	Trailer() metadata.MD
	Send(req interface{}) error
	Receive(res interface{}) error
	CloseSend() error
}

type client struct {
	conn    *grpc.ClientConn
	headers Headers

	grpcreflection.Client
}

// NewClient creates a new gRPC client. It dials to the server specified by addr.
// addr format is the same as the first argument of grpc.Dial.
// If serverName is not empty, it overrides the gRPC server name used to
// verify the hostname on the returned certificates.
// If useReflection is true, the gRPC client enables gRPC reflection.
// If useTLS is true, the gRPC client establishes a secure connection with the server.
//
// The set of cert and certKey enables mutual authentication if useTLS is enabled.
// If one of it is not found, NewClient returns ErrMutualAuthParamsAreNotEnough.
// If useTLS is false, cacert, cert and certKey are ignored.
func NewClient(addr, serverName string, useReflection, useTLS bool, cacert, cert, certKey string, headers map[string][]string) (Client, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(64*1024*1024)))
	if !useTLS {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else { // Enable TLS authentication
		var tlsCfg tls.Config
		if cacert != "" {
			b, err := os.ReadFile(cacert)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read the CA certificate")
			}
			cp := x509.NewCertPool()
			if !cp.AppendCertsFromPEM(b) {
				return nil, errors.New("failed to append the client certificate")
			}
			tlsCfg.RootCAs = cp
		}
		if cert != "" && certKey != "" {
			// Enable mutual authentication
			certificate, err := tls.LoadX509KeyPair(cert, certKey)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read the client certificate")
			}
			tlsCfg.Certificates = append(tlsCfg.Certificates, certificate)
		} else if cert != "" || certKey != "" {
			return nil, ErrMutualAuthParamsAreNotEnough
		}

		creds := credentials.NewTLS(&tlsCfg)
		opts = append(opts, grpc.WithTransportCredentials(creds))

		if serverName != "" {
			opts = append(opts, grpc.WithAuthority(serverName))
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial to gRPC server")
	}

	client := &client{
		conn:    conn,
		headers: Headers{},
	}

	if useReflection {
		client.Client = grpcreflection.NewClient(conn, headers)
	}

	return client, nil
}

func (c *client) Invoke(ctx context.Context, fqrn string, req, res interface{}) (header, trailer metadata.MD, _ error) {
	logger.Scriptln(func() []interface{} {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			return nil
		}
		return []interface{}{md}
	})

	endpoint, err := fqrnToEndpoint(fqrn)
	if err != nil {
		return nil, nil, err
	}
	loggingRequest(req)
	wakeUpClientConn(c.conn)
	opts := []grpc.CallOption{grpc.Header(&header), grpc.Trailer(&trailer)}
	err = c.conn.Invoke(ctx, endpoint, req, res, opts...)
	return header, trailer, err
}

func (c *client) Close(ctx context.Context) error {
	doneCh := make(chan error)
	go func() {
		var result error
		if c.Client != nil {
			c.Client.Reset()
		}
		if err := c.conn.Close(); err != nil {
			result = multierror.Append(result, errors.Wrap(err, "failed to close gRPC client"))
		}
		doneCh <- result
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-doneCh:
		return err
	}
}

func (c *client) Header() Headers {
	return c.headers
}

type clientStream struct {
	cs grpc.ClientStream
}

func (s *clientStream) Header() (metadata.MD, error) {
	return s.cs.Header()
}

func (s *clientStream) Trailer() metadata.MD {
	return s.cs.Trailer()
}

func (s *clientStream) Send(req interface{}) error {
	loggingRequest(req)
	return s.cs.SendMsg(req)
}

func (s *clientStream) CloseAndReceive(res interface{}) error {
	if err := s.cs.CloseSend(); err != nil {
		return errors.Wrap(err, "failed to close client stream")
	}

	err := s.cs.RecvMsg(res)
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed to close and receive response")
	}
	return nil
}

func (c *client) NewClientStream(ctx context.Context, streamDesc *grpc.StreamDesc, fqrn string) (ClientStream, error) {
	endpoint, err := fqrnToEndpoint(fqrn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert fqrn to endpoint")
	}
	wakeUpClientConn(c.conn)
	cs, err := c.conn.NewStream(ctx, streamDesc, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to instantiate gRPC stream")
	}
	return &clientStream{cs}, nil
}

type serverStream struct {
	*clientStream
}

func (s *serverStream) Receive(res interface{}) error {
	return s.cs.RecvMsg(res)
}

func (c *client) NewServerStream(ctx context.Context, streamDesc *grpc.StreamDesc, fqrn string) (ServerStream, error) {
	s, err := c.NewClientStream(ctx, streamDesc, fqrn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create server stream")
	}
	return &serverStream{s.(*clientStream)}, nil
}

type bidiStream struct {
	s *serverStream
}

func (s *bidiStream) Header() (metadata.MD, error) {
	return s.s.Header()
}

func (s *bidiStream) Trailer() metadata.MD {
	return s.s.Trailer()
}

func (s *bidiStream) Send(req interface{}) error {
	loggingRequest(req)
	return s.s.cs.SendMsg(req)
}

func (s *bidiStream) Receive(res interface{}) error {
	return s.s.cs.RecvMsg(res)
}

func (s *bidiStream) CloseSend() error {
	return s.s.cs.CloseSend()
}

func (c *client) NewBidiStream(ctx context.Context, streamDesc *grpc.StreamDesc, fqrn string) (BidiStream, error) {
	s, err := c.NewServerStream(ctx, streamDesc, fqrn)
	if err != nil {
		return nil, err
	}
	return &bidiStream{s.(*serverStream)}, nil
}

// fqrnToEndpoint converts FullQualifiedRPCName to endpoint
//
// e.g.
//
//	pkg_name.svc_name.rpc_name -> /pkg_name.svc_name/rpc_name
func fqrnToEndpoint(fqrn string) (string, error) {
	sp := strings.Split(fqrn, ".")
	// FQRN should contain at least service and rpc name.
	if len(sp) < 2 {
		return "", errors.New("invalid FQRN format")
	}

	return fmt.Sprintf("/%s/%s", strings.Join(sp[:len(sp)-1], "."), sp[len(sp)-1]), nil
}

func wakeUpClientConn(conn *grpc.ClientConn) {
	if conn.GetState() == connectivity.TransientFailure {
		conn.ResetConnectBackoff()
	}
}

func loggingRequest(req interface{}) {
	logger.Scriptln(func() []interface{} {
		b, err := json.MarshalIndent(&req, "", "  ")
		if err != nil {
			return nil
		}
		return []interface{}{"request:\n" + string(b)}
	})
}
