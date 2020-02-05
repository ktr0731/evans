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
	ErrUnknownMessageName = errors.New("unknown message name")
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

	// TypeDescriptor returns the descriptor of a type specified by msgName.
	// Message name should be fully-qualified (the form of <package>.<message> in Protocol Buffers3).
	// The returned descriptor depends to an codec such that Protocol Buffers.
	// TypeDescriptor may returns these errors:
	//
	//   - ErrUnknownMessage: msgName is not loaded.
	//
	TypeDescriptor(msgName string) (interface{}, error)
}

// FullyQualifiedMessageName returns the fully qualified name joined with '.'.
// pkgName is an optional value, but msgName is not.
func FullyQualifiedMessageName(pkgName, msgName string) (string, error) {
	if msgName == "" {
		return "", errors.New("msgName should not be empty")
	}
	if pkgName == "" {
		return msgName, nil
	}
	return strings.Join([]string{pkgName, msgName}, "."), nil
}
