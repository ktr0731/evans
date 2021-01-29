// Package grpcreflection provides gRPC reflection client.
// Currently, gRPC reflection depends on Protocol Buffers, so we split this package from grpc package.
package grpcreflection

import (
	"context"
	"strings"

	"github.com/jhump/protoreflect/desc"
	gr "github.com/jhump/protoreflect/grpcreflect"
	"github.com/ktr0731/grpc-web-go-client/grpcweb"
	"github.com/ktr0731/grpc-web-go-client/grpcweb/grpcweb_reflection_v1alpha"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
)

// ServiceName represents the gRPC reflection service name.
const ServiceName = "grpc.reflection.v1alpha.ServerReflection"

var ErrTLSHandshakeFailed = errors.New("TLS handshake failed")

// Client defines gRPC reflection client.
type Client interface {
	// ListPackages lists file descriptors from the gRPC reflection server.
	// ListPackages returns these errors:
	//   - ErrTLSHandshakeFailed: TLS misconfig.
	ListPackages() ([]*desc.FileDescriptor, error)
	// Reset clears internal states of Client.
	Reset()
}

type client struct {
	client *gr.Client
}

func getCtx(headers map[string][]string) context.Context {
	return metadata.NewOutgoingContext(context.Background(), metadata.MD(headers))
}

// NewClient returns an instance of gRPC reflection client for gRPC protocol.
func NewClient(conn grpc.ClientConnInterface, headers map[string][]string) Client {
	return &client{
		client: gr.NewClient(getCtx(headers), grpc_reflection_v1alpha.NewServerReflectionClient(conn)),
	}
}

// NewWebClient returns an instance of gRPC reflection client for gRPC-Web protocol.
func NewWebClient(conn *grpcweb.ClientConn, headers map[string][]string) Client {
	return &client{
		client: gr.NewClient(getCtx(headers), grpcweb_reflection_v1alpha.NewServerReflectionClient(conn)),
	}
}

func (c *client) ListPackages() ([]*desc.FileDescriptor, error) {
	ssvcs, err := c.client.ListServices()
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
		return nil, errors.Wrap(err, "failed to list services from reflecton enabled gRPC server")
	}

	fds := make([]*desc.FileDescriptor, 0, len(ssvcs))
	for _, s := range ssvcs {
		svc, err := c.client.ResolveService(s)
		if err != nil {
			if gr.IsElementNotFoundError(err) {
				// Service doesn't expose the ServiceDescriptor, skip.
				continue
			}
			return nil, errors.Wrapf(err, "failed to resolve service '%s'", s)
		}

		fds = append(fds, svc.GetFile())
	}

	return fds, nil
}

func (c *client) Reset() {
	c.client.Reset()
}
