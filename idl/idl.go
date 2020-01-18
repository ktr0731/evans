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

	"github.com/ktr0731/evans/grpc"
)

var (
	ErrPackageUnselected = errors.New("package unselected")
	ErrServiceUnselected = errors.New("service unselected")

	ErrUnknownPackageName = errors.New("unknown package name")
	ErrUnknownServiceName = errors.New("unknown service name")
	ErrUnknownRPCName     = errors.New("unknown RPC name")
)

// EmptyPackage indicates it is an empty package (there is no package specifier).
// Package specifier only allows letters and '_', so using ' is name safe.
// See https://developers.google.com/protocol-buffers/docs/reference/proto3-spec#package
const EmptyPackage = "''"

// Spec represents the interface specification from loaded IDL files.
type Spec interface {
	// PackageNames returns all package names. The returned slice is ordered by ascending order.
	PackageNames() []string

	// ServiceNames returns all service names belongs to the passed package name pkgName.
	// ServiceNames may return these errors:
	//
	//   - ErrPackageUnselected: pkgName is empty.
	//   - ErrUnknownPackageName: pkgName is not contained to PackageNames().
	//
	ServiceNames(pkgName string) ([]string, error)

	// RPCs returns all RPC names belongs to the passed service name svcName.
	// RPCs may return these errors:
	//
	//   - ErrPackageUnselected: pkgName is empty.
	//   - ErrServiceUnselected: svcName is empty.
	//   - ErrUnknownPackageName: pkgName is not contained to PackageNames().
	//   - ErrUnknownServiceName: svcName is not contained to ServiceNames().
	//
	RPCs(pkgName, svcName string) ([]*grpc.RPC, error)

	// RPC returns the RPC that is specified by pkgName, svcName and rpcName.
	// RPC may return these errors:
	//
	//   - ErrPackageUnselected: pkgName is empty.
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
	//   - ErrPackageUnselected: pkgName is empty.
	//
	TypeDescriptor(pkgName, msgName string) (interface{}, error)
}
