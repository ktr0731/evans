// Package grpcreflection provides gRPC reflection client.
// Currently, gRPC reflection depends on Protocol Buffers, so we split this package from grpc package.
package grpcreflection

import (
	"context"
	"strings"

	gr "github.com/jhump/protoreflect/grpcreflect"
	"github.com/ktr0731/grpc-web-go-client/grpcweb"
	"github.com/ktr0731/grpc-web-go-client/grpcweb/grpcweb_reflection_v1alpha"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// ServiceName represents the gRPC reflection service name.
const ServiceName = "grpc.reflection.v1alpha.ServerReflection"

var ErrTLSHandshakeFailed = errors.New("TLS handshake failed")

// Client defines gRPC reflection client.
type Client interface {
	// ListServices lists registered service names.
	// ListServices returns these errors:
	//   - ErrTLSHandshakeFailed: TLS misconfig.
	ListServices() ([]string, error)
	// FindSymbol returns the symbol associated with the given name.
	FindSymbol(name string) (protoreflect.Descriptor, error)
	// Reset clears internal states of Client.
	Reset()
}

type client struct {
	resolver *protoregistry.Files
	client   *gr.Client
}

func getCtx(headers map[string][]string) context.Context {
	md := metadata.New(nil)
	for k, v := range headers {
		md.Append(k, v...)
	}
	return metadata.NewOutgoingContext(context.Background(), md)
}

// NewClient returns an instance of gRPC reflection client for gRPC protocol.
func NewClient(conn grpc.ClientConnInterface, headers map[string][]string) Client {
	return &client{
		client:   gr.NewClient(getCtx(headers), grpc_reflection_v1alpha.NewServerReflectionClient(conn)),
		resolver: protoregistry.GlobalFiles,
	}
}

// NewWebClient returns an instance of gRPC reflection client for gRPC-Web protocol.
func NewWebClient(conn *grpcweb.ClientConn, headers map[string][]string) Client {
	return &client{
		client:   gr.NewClient(getCtx(headers), grpcweb_reflection_v1alpha.NewServerReflectionClient(conn)),
		resolver: protoregistry.GlobalFiles,
	}
}

func (c *client) ListServices() ([]string, error) {
	svcs, err := c.client.ListServices()
	if err != nil {
		msg := status.Convert(err).Message()
		// Check whether the error message contains TLS related error.
		// If the server didn't enable TLS, the error message contains the first string.
		// If Evans didn't enable TLS against to the TLS enabled server, the error message contains
		// the second string.
		if strings.Contains(msg, "tls: first record does not look like a TLS handshake") ||
			strings.Contains(msg, "latest connection error: <nil>") {
			return nil, ErrTLSHandshakeFailed
		}
		return nil, errors.Wrap(err, "failed to list services from reflection enabled gRPC server")
	}

	return svcs, nil
}

func (c *client) FindSymbol(name string) (protoreflect.Descriptor, error) {
	fullName := protoreflect.FullName(name)

	d, err := c.resolver.FindDescriptorByName(fullName)
	if err != nil && !errors.Is(err, protoregistry.NotFound) {
		return nil, err
	}
	if err == nil {
		return d, nil
	}

	jfd, err := c.client.FileContainingSymbol(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find file containing symbol")
	}

	// TODO: consider dependencies
	fd, err := protodesc.NewFile(jfd.AsFileDescriptorProto(), c.resolver)
	if err != nil {
		return nil, err
	}

	if err := c.resolver.RegisterFile(fd); err != nil {
		return nil, err
	}

	return c.resolver.FindDescriptorByName(fullName)
}

func (c *client) Reset() {
	c.client.Reset()
}
