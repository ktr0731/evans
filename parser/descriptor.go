package parser

import (
	"github.com/jhump/protoreflect/desc"
)

type fileDescriptorSet []*desc.FileDescriptor
type FileDescriptorSet struct {
	fileDescriptorSet
}

func (d *FileDescriptorSet) GetFile() []*desc.FileDescriptor {
	return d.fileDescriptorSet
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

func (d *FileDescriptorSet) GetServices(pkg string) []*desc.ServiceDescriptor {
	svc := []*desc.ServiceDescriptor{}
	for _, f := range d.GetFile() {
		if f.GetPackage() != pkg {
			continue
		}
		svc = append(svc, f.GetServices()...)
	}

	return svc
}

func (d *FileDescriptorSet) GetMessages(pkg string) []*desc.MessageDescriptor {
	msg := []*desc.MessageDescriptor{}
	for _, f := range d.GetFile() {
		if f.GetPackage() != pkg {
			continue
		}
		msg = append(msg, f.GetMessageTypes()...)
	}

	return msg
}
