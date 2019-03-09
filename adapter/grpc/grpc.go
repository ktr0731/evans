package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-multierror"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/logger"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

type client struct {
	conn *grpc.ClientConn

	*reflectionClient
}

// NewClient creates a new gRPC client. It dials to the server specified by addr.
// addr format is the same as the first argument of grpc.Dial.
// If useReflection is true, the gRPC client enables gRPC reflection.
// If useTLS is true, the gRPC client establishes a secure connection with the server.
//
// The set of cert and certKey enables mutual authentication if useTLS is enabled.
// If one of it is not found, NewClient returns entity.ErrMutualAuthParamsAreNotEnough.
// If useTLS is false, cacert, cert and certKey are ignored.
func NewClient(addr string, useReflection, useTLS bool, cacert, cert, certKey string) (entity.GRPCClient, error) {
	var opts []grpc.DialOption
	if !useTLS {
		opts = append(opts, grpc.WithInsecure())
	} else { // Enable TLS authentication
		var tlsCfg tls.Config
		if cacert != "" {
			b, err := ioutil.ReadFile(cacert)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read the CA certificate")
			}
			cp := x509.NewCertPool()
			if !cp.AppendCertsFromPEM(b) {
				return nil, errors.Wrap(err, "failed to append the client certificate")
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
			opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tlsCfg)))
		} else if cert != "" || certKey != "" {
			return nil, entity.ErrMutualAuthParamsAreNotEnough
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tlsCfg)))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial to gRPC server")
	}
	switch s := conn.GetState(); s {
	case connectivity.TransientFailure:
		return nil, errors.Errorf("connection transient failure, is the gRPC server running?: %s", s)
	case connectivity.Shutdown:
		return nil, errors.Errorf("the gRPC server was closed: %s", s)
	}

	client := &client{
		conn: conn,
	}

	if useReflection {
		client.reflectionClient = newReflectionClient(conn)
	}

	return client, nil
}

func (c *client) Invoke(ctx context.Context, fqrn string, req, res interface{}) error {
	endpoint, err := fqrnToEndpoint(fqrn)
	if err != nil {
		return err
	}
	logger.Scriptln(func() []interface{} {
		b, err := json.MarshalIndent(&req, "", "  ")
		if err != nil {
			return nil
		}
		return []interface{}{"request:\n" + string(b)}
	})
	wakeUpClientConn(c.conn)
	return c.conn.Invoke(ctx, endpoint, req, res)
}

func (c *client) Close(ctx context.Context) error {
	doneCh := make(chan error)
	go func() {
		var result error
		c.reflectionClient.Close()
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

type clientStream struct {
	cs grpc.ClientStream
}

func (s *clientStream) Send(m proto.Message) error {
	return s.cs.SendMsg(m)
}

func (s *clientStream) CloseAndReceive(res *proto.Message) error {
	if err := s.cs.CloseSend(); err != nil {
		return errors.Wrap(err, "failed to close client stream")
	}

	err := s.cs.RecvMsg(*res)
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed to close and receive response")
	}
	return nil
}

func (c *client) NewClientStream(ctx context.Context, rpc entity.RPC) (entity.ClientStream, error) {
	endpoint, err := fqrnToEndpoint(rpc.FQRN())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert fqrn to endpoint")
	}
	cs, err := c.conn.NewStream(ctx, rpc.StreamDesc(), endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to instantiate gRPC stream")
	}
	wakeUpClientConn(c.conn)
	return &clientStream{cs}, nil
}

type serverStream struct {
	*clientStream
}

func (s *serverStream) Receive(res *proto.Message) error {
	return s.cs.RecvMsg(*res)
}

func (c *client) NewServerStream(ctx context.Context, rpc entity.RPC) (entity.ServerStream, error) {
	s, err := c.NewClientStream(ctx, rpc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create server stream")
	}
	return &serverStream{s.(*clientStream)}, nil
}

type bidiStream struct {
	s *serverStream
}

func (s *bidiStream) Send(res proto.Message) error {
	return s.s.cs.SendMsg(res)
}

func (s *bidiStream) Receive(res *proto.Message) error {
	return s.s.cs.RecvMsg(*res)
}

func (s *bidiStream) Close() error {
	return s.s.cs.CloseSend()
}

func (c *client) NewBidiStream(ctx context.Context, rpc entity.RPC) (entity.BidiStream, error) {
	s, err := c.NewServerStream(ctx, rpc)
	if err != nil {
		return nil, err
	}
	return &bidiStream{s.(*serverStream)}, nil
}

// fqrnToEndpoint converts FullQualifiedRPCName to endpoint
//
// e.g.
//	pkg_name.svc_name.rpc_name -> /pkg_name.svc_name/rpc_name
func fqrnToEndpoint(fqrn string) (string, error) {
	sp := strings.Split(fqrn, ".")
	if len(sp) < 3 {
		return "", errors.New("invalid FQRN format")
	}

	return fmt.Sprintf("/%s/%s", strings.Join(sp[:len(sp)-1], "."), sp[len(sp)-1]), nil
}

func wakeUpClientConn(conn *grpc.ClientConn) {
	if conn.GetState() == connectivity.TransientFailure {
		conn.ResetConnectBackoff()
	}
}
