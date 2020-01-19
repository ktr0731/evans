// Package proto implements idl.Spec for Protocol Buffers.
package proto

import (
	"fmt"
	"sort"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/grpc"
	"github.com/ktr0731/evans/grpc/grpcreflection"
	"github.com/ktr0731/evans/idl"
	"github.com/pkg/errors"
)

type spec struct {
	pkgNames []string
	// key: package name, val: service descriptors belong to the package.
	svcDescs map[string][]*desc.ServiceDescriptor
	// key: fully qualified service name, val: method descriptors belong to the service.
	rpcDescs map[string][]*desc.MethodDescriptor
	// key: fully qualified message name, val: the message descriptor.
	msgDescs map[string]*desc.MessageDescriptor
}

func (s *spec) PackageNames() []string {
	return s.pkgNames
}

func (s *spec) ServiceNames(pkgName string) ([]string, error) {
	descs, ok := s.svcDescs[pkgName]
	if !ok {
		// If the service belongs to pkgName is not found and pkgName is empty,
		// it is regarded as package is unselected.
		if pkgName == "" {
			return nil, idl.ErrPackageUnselected
		}
		return nil, idl.ErrUnknownPackageName
	}

	svcNames := make([]string, len(descs))
	for i, d := range descs {
		svcNames[i] = d.GetName()
	}
	return svcNames, nil
}

func (s *spec) RPCs(pkgName, svcName string) ([]*grpc.RPC, error) {
	// Check whether pkgName is a valid package or not.
	_, err := s.ServiceNames(pkgName)
	if err != nil {
		return nil, err
	}

	if svcName == "" {
		return nil, idl.ErrServiceUnselected
	}

	fqsn, err := idl.FullyQualifiedServiceName(pkgName, svcName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get fully-qualified service name")
	}
	rpcDescs, ok := s.rpcDescs[fqsn]
	if !ok {
		return nil, idl.ErrUnknownServiceName
	}

	rpcs := make([]*grpc.RPC, len(rpcDescs))
	for i, d := range rpcDescs {
		rpc, err := s.RPC(pkgName, svcName, d.GetName())
		if err != nil {
			panic(fmt.Sprintf("RPC must not return an error, but got '%s'", err))
		}
		rpcs[i] = rpc
	}
	return rpcs, nil
}

func (s *spec) RPC(pkgName, svcName, rpcName string) (*grpc.RPC, error) {
	// Check whether pkgName is a valid package or not.
	_, err := s.ServiceNames(pkgName)
	if err != nil {
		return nil, err
	}

	if svcName == "" {
		return nil, idl.ErrServiceUnselected
	}

	fqsn, err := idl.FullyQualifiedServiceName(pkgName, svcName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get fully-qualified service name")
	}
	rpcDescs, ok := s.rpcDescs[fqsn]
	if !ok {
		return nil, idl.ErrUnknownServiceName
	}

	for _, d := range rpcDescs {
		if d.GetName() == rpcName {
			return &grpc.RPC{
				Name:               d.GetName(),
				FullyQualifiedName: d.GetFullyQualifiedName(),
				RequestType: &grpc.Type{
					Name:               d.GetInputType().GetName(),
					FullyQualifiedName: d.GetInputType().GetFullyQualifiedName(),
					New: func() (interface{}, error) {
						m := dynamic.NewMessage(d.GetInputType())
						return m, nil
					},
				},
				ResponseType: &grpc.Type{
					Name:               d.GetOutputType().GetName(),
					FullyQualifiedName: d.GetOutputType().GetFullyQualifiedName(),
					New: func() (interface{}, error) {
						m := dynamic.NewMessage(d.GetOutputType())
						return m, nil
					},
				},
				IsServerStreaming: d.IsServerStreaming(),
				IsClientStreaming: d.IsClientStreaming(),
			}, nil
		}
	}
	return nil, idl.ErrUnknownRPCName
}

// TypeDescriptor returns the descriptor of a type.
// The actual type of the returned interface{} is *desc.MessageDescriptor.
func (s *spec) TypeDescriptor(pkgName, msgName string) (interface{}, error) {
	if pkgName == "" {
		return nil, idl.ErrPackageUnselected
	}
	fqtn := pkgName + "." + msgName
	if m, ok := s.msgDescs[fqtn]; ok {
		return m, nil
	}
	return nil, errors.Errorf("no such type '%s'", fqtn)
}

// LoadFiles receives proto file names and import paths like protoc's options.
// Then, LoadFiles parses these files and instantiates a new idl.Spec.
func LoadFiles(importPaths []string, fnames []string) (idl.Spec, error) {
	p := &protoparse.Parser{
		ImportPaths: importPaths,
	}
	fileDescs, err := p.ParseFiles(fnames...)
	if err != nil {
		return nil, errors.Wrap(err, "proto: failed to parse passed proto files")
	}

	// Collect dependency file descriptors
	for _, d := range fileDescs {
		fileDescs = append(fileDescs, d.GetDependencies()...)
	}

	return newSpec(fileDescs), nil
}

// LoadByReflection receives a gRPC reflection client, then tries to instantiate a new idl.Spec by using gRPC reflection.
func LoadByReflection(client grpcreflection.Client) (idl.Spec, error) {
	fileDescs, err := client.ListPackages()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list packages by gRPC reflection")
	}
	return newSpec(fileDescs), nil
}

func newSpec(fds []*desc.FileDescriptor) idl.Spec {
	encounteredPackages := make(map[string]interface{})
	var pkgNames []string
	svcDescs := make(map[string][]*desc.ServiceDescriptor)
	rpcDescs := make(map[string][]*desc.MethodDescriptor)
	msgDescs := make(map[string]*desc.MessageDescriptor)
	for _, f := range fds {
		pkg := f.GetPackage()
		svcDescs[pkg] = append(svcDescs[pkg], f.GetServices()...)
		if _, encountered := encounteredPackages[pkg]; !encountered {
			pkgNames = append(pkgNames, pkg)
			encounteredPackages[pkg] = nil
		}

		for _, s := range f.GetServices() {
			rpcDescs[s.GetFullyQualifiedName()] = append(rpcDescs[s.GetFullyQualifiedName()], s.GetMethods()...)
		}

		for _, m := range f.GetMessageTypes() {
			msgDescs[m.GetFullyQualifiedName()] = m
		}
	}

	sort.Slice(pkgNames, func(i, j int) bool {
		return pkgNames[i] < pkgNames[j]
	})

	return &spec{
		pkgNames: pkgNames,
		svcDescs: svcDescs,
		rpcDescs: rpcDescs,
		msgDescs: msgDescs,
	}
}
