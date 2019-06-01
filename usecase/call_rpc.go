package usecase

import (
	"context"
	"io"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// CallRPC constructs a request with input source such that prompt inputting, stdin or a file. After that, it sends
// the request to the gRPC server and decodes the response body to res.
// Note that req and res must be JSON-decodable structs. The output is written to w.
func CallRPC(ctx context.Context, w io.Writer, rpcName string) error {
	return dm.CallRPC(ctx, w, rpcName)
}
func (m *dependencyManager) CallRPC(ctx context.Context, w io.Writer, rpcName string) error {
	rpc, err := m.spec.RPC(m.state.selectedPackage, m.state.selectedService, rpcName)
	if err != nil {
		return errors.Wrap(err, "failed to get the RPC descriptor")
	}
	newRequest := func() (interface{}, error) {
		req, err := rpc.RequestType.New()
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to instantiate an instance of the request type '%s'",
				rpc.RequestType.FullyQualifiedName)
		}
		err = m.filler.Fill(req)
		if err == io.EOF {
			return nil, io.EOF
		}
		if err != nil {
			return nil, err
		}
		return req, nil
	}
	newResponse := func() (interface{}, error) {
		res, err := rpc.ResponseType.New()
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to instantiate an instance of the response type '%s'",
				rpc.RequestType.FullyQualifiedName)
		}
		return res, nil
	}
	flushResponse := func(res interface{}) error {
		out, err := m.presenter.Format(res, "  ")
		if err != nil {
			return err
		}
		if _, err := io.WriteString(w, out+"\n"); err != nil {
			return err
		}
		return nil
	}

	md := metadata.New(nil)
	for k, v := range m.ListHeaders() {
		md.Append(k, v...)
	}
	ctx = metadata.NewOutgoingContext(ctx, md)

	streamDesc := &gogrpc.StreamDesc{
		StreamName:    rpc.Name,
		ServerStreams: rpc.IsServerStreaming,
		ClientStreams: rpc.IsClientStreaming,
	}
	switch {
	case rpc.IsClientStreaming && rpc.IsServerStreaming:
		stream, err := m.gRPCClient.NewBidiStream(ctx, streamDesc, rpc.FullyQualifiedName)
		if err != nil {
			return errors.Wrapf(err, "failed to create a bidi stream for RPC '%s'", streamDesc.StreamName)
		}

		var eg errgroup.Group
		eg.Go(func() error {
			for {
				res, err := newResponse()
				if err != nil {
					return err
				}
				err = stream.Receive(res)
				switch {
				case err == context.Canceled:
					return nil
				case err == io.EOF:
					return nil
				case err != nil:
					return errors.Wrapf(
						err,
						"failed to receive a response from the server stream '%s'",
						streamDesc.StreamName)
				}
				if err := flushResponse(res); err != nil {
					return err
				}
			}
		})

		eg.Go(func() error {
			for {
				req, err := newRequest()
				if err == io.EOF {
					if err := stream.CloseSend(); err != nil {
						return errors.Wrapf(
							err,
							"failed to close the stream of RPC '%s'",
							streamDesc.StreamName)
					}
					return nil
				}
				if err != nil {
					return err
				}
				if err := stream.Send(req); err != nil {
					return errors.Wrapf(
						err,
						"failed to send a RPC to the client stream '%s'",
						streamDesc.StreamName)
				}
			}
		})

		if err := eg.Wait(); err != nil {
			return errors.Wrap(err, "failed to process bidi streaming RPC")
		}
		return nil

	// Client streaming RPCs are RPC that a client sends several times and server responds once.
	// Client streaming RPCs are processed by the following instruction.
	//
	//   1. Create a new client stream.
	//   2. Create a new request and fill input to it.
	//   3. Send the request to the server.
	//   4. Repeat 1-3 until the filler returns io.EOF.
	//   5. Send the close message and receive the response.
	//   6. Format the response and output it.
	//
	case rpc.IsClientStreaming:
		stream, err := m.gRPCClient.NewClientStream(ctx, streamDesc, rpc.FullyQualifiedName)
		if err != nil {
			return errors.Wrapf(
				err,
				"failed to create a new client stream for RPC '%s'",
				streamDesc.StreamName)
		}
		for {
			req, err := newRequest()
			if err == io.EOF {
				res, err := newResponse()
				if err != nil {
					return err
				}
				if err := stream.CloseAndReceive(res); err != nil {
					return errors.Wrapf(
						err,
						"failed to close the stream of RPC '%s'",
						streamDesc.StreamName)
				}
				if err := flushResponse(res); err != nil {
					return err
				}
				return nil
			}
			if err != nil {
				return err
			}
			if err := stream.Send(req); err != nil {
				return errors.Wrapf(
					err,
					"failed to send a RPC to the client stream '%s'",
					streamDesc.StreamName)
			}
		}

	// Server streaming RPCs are RPC that a client sends once and server responds several times.
	// Server streaming RPCs are processed by the following instruction.
	//
	//   1. Create a new request and fill input to it.
	//   2. Send the request to the server.
	//   3. Call Receive to receive server responses.
	//   4. Format a received response and output it.
	//   5. If io.EOF received, finish the RPC connection.
	//
	case rpc.IsServerStreaming:
		stream, err := m.gRPCClient.NewServerStream(ctx, streamDesc, rpc.FullyQualifiedName)
		if err != nil {
			return errors.Wrapf(
				err,
				"failed to create a new server stream for RPC '%s'",
				streamDesc.StreamName)
		}
		req, err := newRequest()
		if err != nil {
			return err
		}
		if err := stream.Send(req); err != nil {
			return errors.Wrapf(
				err,
				"failed to send a RPC to the server stream '%s'",
				streamDesc.StreamName)
		}

		for {
			res, err := newResponse()
			if err != nil {
				return err
			}
			err = stream.Receive(res)
			switch {
			case err == context.Canceled:
				return nil
			case err == io.EOF:
				return nil
			case err != nil:
				return errors.Wrapf(
					err,
					"failed to receive a response from the server stream '%s'",
					streamDesc.StreamName)
			}
			if err := flushResponse(res); err != nil {
				return err
			}
		}

	// If both of rpc.IsServerStreaming and rpc.IsClientStreaming are nil, it means its RPC is an unary RPC.
	// Unary RPCs are processed by the following instruction.
	//
	//   1. Create a new request and fill input to it.
	//   2. Create a new response.
	//   3. Invoke the RPC with the request and decode response to 2's instance.
	//   4. Format the response.
	//
	default:
		req, err := newRequest()
		if err != nil {
			return err
		}
		res, err := newResponse()
		if err != nil {
			return err
		}
		if err := m.gRPCClient.Invoke(ctx, rpc.FullyQualifiedName, req, res); err != nil {
			return errors.Wrap(err, "failed to send a request")
		}
		if err := flushResponse(res); err != nil {
			return err
		}
		return nil
	}
}
