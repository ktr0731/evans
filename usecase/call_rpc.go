package usecase

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/grpc"
	"github.com/ktr0731/evans/idl/proto"

	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/logger"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ErrorCode represents an application error code.
type ErrorCode int

// String implements fmt.Stringer.
func (e ErrorCode) String() string {
	return codes.Code(e).String()
}

type gRPCError struct {
	*status.Status
}

func (e *gRPCError) Unwrap() error {
	return e.Status.Err()
}

func (e *gRPCError) Error() string {
	return e.Status.Err().Error()
}

func (e *gRPCError) Code() ErrorCode {
	return ErrorCode(e.Status.Code())
}

// CallRPC constructs a request with input source such that prompt inputting, stdin or a file. After that, it sends
// the request to the gRPC server and decodes the response body to res.
// Note that req and res must be JSON-decodable structs. The output is written to w.
func CallRPC(ctx context.Context, w io.Writer, rpcName string) error {
	return dm.CallRPC(ctx, w, rpcName, false, dm.filler)
}
func (m *dependencyManager) CallRPC(ctx context.Context, w io.Writer, rpcName string, rerunPrevious bool, filler fill.Filler) error {
	fqsn := proto.FullyQualifiedServiceName(m.state.selectedPackage, m.state.selectedService)
	rpc, err := m.spec.RPC(fqsn, rpcName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get the RPC descriptor for: %s", rpcName))
	}
	newRequest := func() (interface{}, error) {
		req := rpc.RequestType.New()
		if !rerunPrevious {
			err = filler.Fill(req)
			if errors.Is(err, io.EOF) {
				return nil, io.EOF
			}
			if err != nil {
				return nil, err
			}
			if err := m.updateRPCCallState(rpcName, req, rpc); err != nil {
				return nil, err
			}
			return req, nil
		}
		if err = m.getPreviousRPCRequest(rpcName, req); err != nil {
			return nil, err
		}
		return req, nil
	}
	newResponse := func() interface{} {
		return rpc.ResponseType.New()
	}
	flushHeader := func(header metadata.MD) {
		m.responseFormatter.FormatHeader(header)
	}
	flushResponse := func(res interface{}) error {
		return m.responseFormatter.FormatMessage(res)
	}
	flushTrailer := func(status *status.Status, trailer metadata.MD) error {
		return m.responseFormatter.FormatTrailer(status, trailer)
	}
	flushDone := func() error {
		return m.responseFormatter.Done()
	}
	flushAll := func(status *status.Status, header, trailer metadata.MD, res interface{}) error {
		flushHeader(header)
		if err := flushResponse(res); err != nil {
			return err
		}
		if err := flushTrailer(status, trailer); err != nil {
			return err
		}
		return flushDone()
	}

	parseDuration := func(duration string) (time.Duration, error) {
		replacer := strings.NewReplacer("n", "ns", "u", "us", "m", "ms", "S", "s", "M", "m", "H", "h")
		duration = replacer.Replace(duration)
		timeout, err := time.ParseDuration(duration)
		if err != nil {
			return 0, errors.Wrapf(err, "malformed grpc-timeout header")
		}
		return timeout, err
	}

	enhanceContext := func(ctx context.Context) (context.Context, context.CancelFunc, error) {
		md := metadata.New(nil)
		for k, v := range m.ListHeaders() {
			md.Append(k, v...)
		}

		ctx = metadata.NewOutgoingContext(ctx, md)
		values := md.Get("grpc-timeout")
		if len(values) == 0 {
			return ctx, func() {}, nil
		}

		timeout, err := parseDuration(values[len(values)-1])
		if err != nil {
			return nil, func() {}, err
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)

		return ctx, cancel, nil
	}

	streamDesc := &gogrpc.StreamDesc{
		StreamName:    rpc.Name,
		ServerStreams: rpc.IsServerStreaming,
		ClientStreams: rpc.IsClientStreaming,
	}

	switch {
	case rpc.IsClientStreaming && rpc.IsServerStreaming:
		ctx, cancel, err := enhanceContext(ctx)
		if err != nil {
			cancel()
			return errors.Wrap(err, "failed to enhance context with metadata")
		}

		stream, err := m.gRPCClient.NewBidiStream(ctx, streamDesc, rpc.FullyQualifiedName)
		if err != nil {
			cancel()
			return errors.Wrapf(err, "failed to create a bidi stream for RPC '%s'", streamDesc.StreamName)
		}

		var (
			eg                                errgroup.Group
			writeHeaderOnce, writeTrailerOnce sync.Once
		)
		eg.Go(func() error {
			for {
				res := newResponse()
				stat, err := handleGRPCResponseError(stream.Receive(res))
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return nil
					}
					if errors.Is(err, io.EOF) {
						writeTrailerOnce.Do(func() {
							if err := flushTrailer(status.New(codes.OK, ""), stream.Trailer()); err != nil {
								logger.Printf("failed to call Done: %s", err)
							}
							if err := flushDone(); err != nil {
								logger.Printf("failed to call Done: %s", err)
							}
						})
						return nil
					}
					return errors.Wrapf(err, "failed to receive a response from the server stream '%s'", streamDesc.StreamName)
				}

				if stat != nil {
					defer func(stat *status.Status) {
						writeTrailerOnce.Do(func() {
							if err := flushTrailer(stat, stream.Trailer()); err != nil {
								logger.Printf("failed to call Done: %s", err)
							}
							if err := flushDone(); err != nil {
								logger.Printf("failed to call Done: %s", err)
							}
						})
					}(stat)
				}

				var whErr error
				writeHeaderOnce.Do(func() {
					header, err := stream.Header()
					if err != nil {
						whErr = errors.Wrap(err, "failed to get header metadata")
					}
					flushHeader(header)
				})
				if whErr != nil {
					return whErr
				}

				if stat.Code() != codes.OK {
					return &gRPCError{stat}
				}

				if err := flushResponse(res); err != nil {
					return err
				}
			}
		})

		eg.Go(func() error {
			for {
				req, err := newRequest()
				if errors.Is(err, io.EOF) {
					if err := stream.CloseSend(); err != nil {
						return errors.Wrapf(err, "failed to close the stream of RPC '%s'", streamDesc.StreamName)
					}
					return nil
				}
				if err != nil {
					return err
				}
				err = stream.Send(req)
				if errors.Is(err, io.EOF) {
					return nil
				}
				if err != nil {
					return errors.Wrapf(err, "failed to send a RPC to the client stream '%s'", streamDesc.StreamName)
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
		ctx, cancel, err := enhanceContext(ctx)
		if err != nil {
			cancel()
			return errors.Wrap(err, "failed to enhance context with metadata")
		}

		stream, err := m.gRPCClient.NewClientStream(ctx, streamDesc, rpc.FullyQualifiedName)
		if err != nil {
			cancel()
			return errors.Wrapf(err, "failed to create a new client stream for RPC '%s'", streamDesc.StreamName)
		}

		for {
			req, err := newRequest()

			if errors.Is(err, io.EOF) {
				res := newResponse()
				stat, err := handleGRPCResponseError(stream.CloseAndReceive(res))
				if err != nil {
					return errors.Wrapf(err, "failed to close the stream of RPC '%s'", streamDesc.StreamName)
				}

				// gRPC error. Treat as a normal response.

				header, err := stream.Header()
				if err != nil {
					return errors.Wrap(err, "failed to get header metadata")
				}

				if stat.Code() != codes.OK {
					res = nil
				}

				err = flushAll(stat, header, stream.Trailer(), res)
				if err != nil {
					return err
				}

				if stat.Code() != codes.OK {
					return &gRPCError{stat}
				}
				return nil
			}
			if err != nil {
				return err
			}
			if err := stream.Send(req); err != nil {
				return errors.Wrapf(err, "failed to send a RPC to the client stream '%s'", streamDesc.StreamName)
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
		req, err := newRequest()
		if err != nil {
			return err
		}

		ctx, cancel, err := enhanceContext(ctx)
		if err != nil {
			cancel()
			return errors.Wrap(err, "failed to enhance context with metadata")
		}

		stream, err := m.gRPCClient.NewServerStream(ctx, streamDesc, rpc.FullyQualifiedName)
		if err != nil {
			cancel()
			return errors.Wrapf(err, "failed to create a new server stream for RPC '%s'", streamDesc.StreamName)
		}

		if err := stream.Send(req); err != nil {
			return errors.Wrapf(err, "failed to send a RPC to the server stream '%s'", streamDesc.StreamName)
		}

		var writeHeaderOnce, writeTrailerOnce sync.Once

		for {
			res := newResponse()
			stat, err := handleGRPCResponseError(stream.Receive(res))
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				if errors.Is(err, io.EOF) {
					return nil
				}
				return errors.Wrapf(err, "failed to receive a response from the server stream '%s'", streamDesc.StreamName)
			}

			// Trailer is now available.
			defer func(stat *status.Status) {
				writeTrailerOnce.Do(func() {
					if err := flushTrailer(stat, stream.Trailer()); err != nil {
						logger.Printf("failed to call Done: %s", err)
					}
					if err := flushDone(); err != nil {
						logger.Printf("failed to call Done: %s", err)
					}
				})
			}(stat)

			var whErr error
			writeHeaderOnce.Do(func() {
				header, err := stream.Header()
				if err != nil {
					whErr = errors.Wrap(err, "failed to get header metadata")
				}
				flushHeader(header)
			})
			if whErr != nil {
				return whErr
			}

			if stat.Code() != codes.OK {
				return &gRPCError{stat}
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

		ctx, cancel, err := enhanceContext(ctx)
		if err != nil {
			cancel()
			return errors.Wrap(err, "failed to enhance context with metadata")
		}

		res := newResponse()
		header, trailer, err := m.gRPCClient.Invoke(ctx, rpc.FullyQualifiedName, req, res)
		stat, err := handleGRPCResponseError(err)
		if err != nil {
			cancel()
			return errors.Wrap(err, "failed to send a request")
		}

		if stat.Code() != codes.OK {
			res = nil
		}

		err = flushAll(stat, header, trailer, res)
		if err != nil {
			return err
		}

		if stat.Code() != codes.OK {
			return &gRPCError{stat}
		}
		return nil
	}
}

// Gets a request with the body containing the payload of its previous RPC
// Only RPCs that are repeatable by definition will have their previous requests returned.
func (m *dependencyManager) getPreviousRPCRequest(rpcName string, req interface{}) error {
	id := rpcIdentifier(rpcName)
	if !m.state.rpcCallState[id].repeatable {
		return errors.New("cannot rerun previous RPC as client streaming RPCs are not supported")
	}
	previousReqBytes := m.state.rpcCallState[id].reqPayload
	if previousReqBytes == nil {
		return errors.New(fmt.Sprintf("No previous request body exists for RPC: %s. Please issue a normal request.", rpcName))
	}
	message := req.(*dynamic.Message)
	err := message.Unmarshal(previousReqBytes)
	if err != nil {
		return errors.Wrap(err, "Error while unmarshalling request for RPC [%s]. Please run without the -r option")
	}
	return nil
}

// Updates the last call state for the given RPC. This is done by serializing
// the request payload and store it into the state buffer indexed by the rpcName
// The RPC is repeatable only if it is not a client streaming rpc.
func (m *dependencyManager) updateRPCCallState(rpcName string, req interface{}, rpc *grpc.RPC) error {
	message := req.(*dynamic.Message)
	reqBytes, err := message.Marshal()
	if err != nil {
		return err
	}
	if m.state.rpcCallState == nil {
		m.state.rpcCallState = make(map[rpcIdentifier]callState)
	}
	m.state.rpcCallState[rpcIdentifier(rpcName)] = callState{
		reqPayload: reqBytes,
		repeatable: !rpc.IsClientStreaming,
	}
	return nil
}

type interactiveFiller struct {
	fillFunc func(v interface{}) error
}

func (f *interactiveFiller) Fill(v interface{}) error {
	return f.fillFunc(v)
}

func CallRPCInteractively(ctx context.Context, w io.Writer, rpcName string, digManually, bytesFromFile bool, rerunPrevious bool) error {
	return dm.CallRPCInteractively(ctx, w, rpcName, digManually, bytesFromFile, rerunPrevious)
}

func (m *dependencyManager) CallRPCInteractively(ctx context.Context, w io.Writer, rpcName string, digManually, bytesFromFile, rerunPrevious bool) error {
	return m.CallRPC(ctx, w, rpcName, rerunPrevious, &interactiveFiller{
		fillFunc: func(v interface{}) error {
			return m.interactiveFiller.Fill(v, fill.InteractiveFillerOpts{DigManually: digManually, BytesFromFile: bytesFromFile})
		},
	})
}

func handleGRPCResponseError(err error) (*status.Status, error) {
	stat, ok := status.FromError(errors.Cause(err))
	if !ok {
		return nil, err
	}
	return stat, nil
}
