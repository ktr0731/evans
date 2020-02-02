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

	ErrInvalidRPCName = errors.New("invalid RPC name")
)

// Spec represents the interface specification from loaded IDL files.
type Spec interface {
	// PackageNames returns all package names. The returned slice is ordered by ascending order.
	PackageNames() []string

	// ServiceNames returns all service names belongs to the passed package name pkgName.
	// ServiceNames may return these errors:
	//
	//   - ErrPackageUnselected: pkgName is empty and there are no IDL files that don't have a package name.
	//   - ErrUnknownPackageName: pkgName is not contained to PackageNames().
	//
	ServiceNames(pkgName string) ([]string, error)

	// RPCs returns all RPC names belongs to the passed service name svcName.
	// RPCs may return these errors:
	//
	//   - ErrPackageUnselected: pkgName is empty and there are no IDL files that don't have a package name.
	//   - ErrServiceUnselected: svcName is empty.
	//   - ErrUnknownPackageName: pkgName is not contained to PackageNames().
	//   - ErrUnknownServiceName: svcName is not contained to ServiceNames().
	//
	RPCs(pkgName, svcName string) ([]*grpc.RPC, error)

	// RPC returns the RPC that is specified by pkgName, svcName and rpcName.
	// RPC may return these errors:
	//
	//   - ErrPackageUnselected: pkgName is empty and there are no IDL files that don't have a package name.
	//   - ErrServiceUnselected: svcName is empty.
	//   - ErrUnknownPackageName: pkgName is not contained to PackageNames().
	//   - ErrUnknownServiceName: svcName is not contained to ServiceNames().
	//   - ErrUnknownRPCName: rpcName is not contained to RPCs().
	//
	RPC(pkgName, svcName, rpcName string) (*grpc.RPC, error)

	// TypeDescriptor returns the descriptor of a type specified by pkgName and msgName.
	// The returned descriptor depends to an codec such that Protocol Buffers.
	// TypeDescriptor may returns these errors:
	//
	//   - ErrPackageUnselected: pkgName is empty and there are no IDL files that don't have a package name.
	//
	TypeDescriptor(pkgName, msgName string) (interface{}, error)
}

// FullyQualifiedServiceName returns the fully qualified name joined with '.'.
// pkgName is an optional value, but svcName is not.
func FullyQualifiedServiceName(pkgName, svcName string) (string, error) {
	if svcName == "" {
		return "", ErrServiceUnselected
	}
	if pkgName == "" {
		return svcName, nil
	}
	return strings.Join([]string{pkgName, svcName}, "."), nil
}

// FullyQualifiedRPCName returns the fully qualified RPC joined with '.'.
// pkgName is an optional value, but others are not.
func FullyQualifiedRPCName(pkgName, svcName, rpcName string) (string, error) {
	s, err := FullyQualifiedServiceName(pkgName, svcName)
	if err != nil {
		return "", err
	}
	if rpcName == "" {
		return "", ErrInvalidRPCName
	}
	return strings.Join([]string{s, rpcName}, "."), nil
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
