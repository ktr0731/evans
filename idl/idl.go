// Package idl represents an Interface Definition Language (IDL) for gRPC.
// In general, Protocol Buffers is used as the IDL. However, it is possible to use another IDLs such that FlatBuffers.
//
// Note that IDLs are independent from encoding.Codec of gRPC.
// It is possible to use a different languages for interface defining and encoding
// (e.g. use Protocol Buffers as an IDL, and use JSON as a codec).
//
// Currently, Evans only supports Protocol Buffers as an IDL.
package idl

import (
	"errors"
	"strings"

	"github.com/ktr0731/evans/grpc"
)

var (
	ErrPackageUnselected = errors.New("package unselected")
	ErrServiceUnselected = errors.New("service unselected")

	ErrUnknownPackageName = errors.New("unknown package name")
	ErrUnknownServiceName = errors.New("unknown service name")
	ErrUnknownRPCName     = errors.New("unknown RPC name")
	ErrUnknownSymbol      = errors.New("unknown symbol")
)

// Spec represents the interface specification from loaded IDL files.
type Spec interface {
	// ServiceNames returns all service names the spec loaded.
	// Service names are fully-qualified (the form of <package>.<service> in Protocol Buffers3).
	// The returned slice is ordered by ascending order.
	ServiceNames() []string

	// RPCs returns all RPC names belongs to the passed fully-qualified service name svcName.
	// RPCs may return these errors:
	//
	//   - ErrServiceUnselected: svcName is empty.
	//   - ErrUnknownServiceName: svcName is not contained to ServiceNames().
	//
	RPCs(svcName string) ([]*grpc.RPC, error)

	// RPC returns the RPC that is specified by svcName and rpcName.
	// RPC may return these errors:
	//
	//   - ErrServiceUnselected: svcName is empty.
	//   - ErrUnknownServiceName: svcName is not contained to ServiceNames().
	//   - ErrUnknownRPCName: rpcName is not contained to RPCs().
	//
	RPC(svcName, rpcName string) (*grpc.RPC, error)

	// ResolveSymbol returns the descriptor of a symbol.
	// The symbol should be fully-qualified (the form of <package>.<message> in Protocol Buffers3).
	// The returned descriptor depends to an codec such that Protocol Buffers.
	// ResolveSymbol may returns these errors:
	//
	//   - ErrUnknownSymbol: symbol is not loaded.
	//
	ResolveSymbol(symbol string) (interface{}, error)

	// FormatDescriptor formats v according to its IDL type.
	FormatDescriptor(v interface{}) (string, error)
}

// FullyQualifiedMethodName returns the fully-qualified method joined with '.'.
func FullyQualifiedMethodName(fqsn, methodName string) (string, error) {
	if fqsn == "" {
		return "", errors.New("fqsn should not be empty")
	}
	if methodName == "" {
		return "", errors.New("methodName should not be empty")
	}
	return strings.Join([]string{fqsn, methodName}, "."), nil
}
