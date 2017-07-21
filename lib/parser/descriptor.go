package parser

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

type FileDescriptorSet struct {
	*descriptor.FileDescriptorSet
}

func (d *FileDescriptorSet) GetPackages() []string {
	packages := make([]string, 0, len(d.GetFile()))
	isEncountered := make(map[string]bool, len(d.GetFile()))

	for _, f := range d.GetFile() {
		if !isEncountered[f.GetPackage()] {
			isEncountered[f.GetPackage()] = true
			packages = append(packages, f.GetPackage())
		}
	}

	return packages
}

func (d *FileDescriptorSet) GetServices(pkg string) []*descriptor.ServiceDescriptorProto {
	svc := []*descriptor.ServiceDescriptorProto{}
	for _, f := range d.GetFile() {
		if f.GetPackage() != pkg {
			continue
		}
		svc = append(svc, f.GetService()...)
	}

	return svc
}

func (d *FileDescriptorSet) GetMessages(pkg string) []*descriptor.DescriptorProto {
	msg := []*descriptor.DescriptorProto{}
	for _, f := range d.GetFile() {
		if f.GetPackage() != pkg {
			continue
		}
		msg = append(msg, f.GetMessageType()...)
	}

	return msg
}
