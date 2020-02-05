// Package proto implements idl.Spec for Protocol Buffers.
package proto

import (
	"fmt"
	"sort"
	"strings"

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
	// Loaded service descriptors.
	svcDescs []*desc.ServiceDescriptor
	// key: fully qualified service name, val: method descriptors belong to the service.
	rpcDescs map[string][]*desc.MethodDescriptor
	// key: fully qualified message name, val: the message descriptor.
	msgDescs map[string]*desc.MessageDescriptor
}

func (s *spec) PackageNames() []string {
	return s.pkgNames
}

func (s *spec) ServiceNames() []string {
	svcNames := make([]string, len(s.svcDescs))
	for i, d := range s.svcDescs {
		svcNames[i] = d.GetFullyQualifiedName()
	}
	return svcNames
}

func (s *spec) RPCs(svcName string) ([]*grpc.RPC, error) {
	if svcName == "" {
		return nil, idl.ErrServiceUnselected
	}

	rpcDescs, ok := s.rpcDescs[svcName]
	if !ok {
		return nil, idl.ErrUnknownServiceName
	}

	rpcs := make([]*grpc.RPC, len(rpcDescs))
	for i, d := range rpcDescs {
		rpc, err := s.RPC(svcName, d.GetName())
		if err != nil {
			panic(fmt.Sprintf("RPC must not return an error, but got '%s'", err))
		}
		rpcs[i] = rpc
	}
	return rpcs, nil
}

func (s *spec) RPC(svcName, rpcName string) (*grpc.RPC, error) {
	if svcName == "" {
		return nil, idl.ErrServiceUnselected
	}

	rpcDescs, ok := s.rpcDescs[svcName]
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

// TypeDescriptor returns the descriptor of a message.
// The actual type of the returned interface{} is *desc.MessageDescriptor.
func (s *spec) TypeDescriptor(msgName string) (interface{}, error) {
	if m, ok := s.msgDescs[msgName]; ok {
		return m, nil
	}
	return nil, idl.ErrUnknownMessageName
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
	var (
		encounteredPkgs = make(map[string]interface{})
		encounteredSvcs = make(map[string]interface{})
		pkgNames        []string
		svcDescs        []*desc.ServiceDescriptor
		rpcDescs        = make(map[string][]*desc.MethodDescriptor)
		msgDescs        = make(map[string]*desc.MessageDescriptor)
	)
	for _, f := range fds {
		if _, encountered := encounteredPkgs[f.GetPackage()]; !encountered {
			pkgNames = append(pkgNames, f.GetPackage())
			encounteredPkgs[f.GetPackage()] = nil
		}
		for _, svc := range f.GetServices() {
			fqsn := svc.GetFullyQualifiedName()
			if _, encountered := encounteredSvcs[fqsn]; !encountered {
				svcDescs = append(svcDescs, svc)
				encounteredSvcs[fqsn] = nil
			}
			rpcDescs[fqsn] = append(rpcDescs[fqsn], svc.GetMethods()...)
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

// FullyQualifiedServiceName returns the fully-qualified service name.
func FullyQualifiedServiceName(pkg, svc string) string {
	var s []string
	if pkg != "" {
		s = []string{pkg, svc}
	} else {
		s = []string{svc}
	}
	return strings.Join(s, ".")
}

// FullyQualifiedMessageName returns the fully-qualified message name.
func FullyQualifiedMessageName(pkg, msg string) string {
	var s []string
	if pkg != "" {
		s = []string{pkg, msg}
	} else {
		s = []string{msg}
	}
	return strings.Join(s, ".")
}
