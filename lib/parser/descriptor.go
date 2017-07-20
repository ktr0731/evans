package parser

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/lycoris0731/evans/lib/model"
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

func (d *FileDescriptorSet) GetServices(pack string) model.Services {
	var services model.Services
	for _, f := range d.GetFile() {
		if f.GetPackage() != pack {
			continue
		}

		for _, proto := range f.GetService() {
			services = append(services, model.NewService(proto))
		}
	}

	return services
}

func (d *FileDescriptorSet) GetMessages(pack string) model.Messages {
	var messages model.Messages
	for _, f := range d.GetFile() {
		if f.GetPackage() != pack {
			continue
		}

		for _, proto := range f.GetMessageType() {
			messages = append(messages, model.NewMessage(proto))
		}
	}
	return messages
}
