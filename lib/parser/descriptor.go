package parser

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/lycoris0731/evans/lib/model"
)

type FileDescriptorSet struct {
	*descriptor.FileDescriptorSet
}

func (d *FileDescriptorSet) GetPackages() model.Packages {
	var packages model.Packages
	isEncountered := make(map[string]bool, len(d.GetFile()))
	for _, f := range d.GetFile() {
		if !isEncountered[f.GetName()] {
			isEncountered[f.GetName()] = true
			packages = append(packages, f.GetPackage())
		}
	}
	return packages
}

func (d *FileDescriptorSet) GetServices() model.Services {
	// TODO: Optimization
	var services model.Services
	for _, f := range d.GetFile() {
		for _, proto := range f.GetService() {
			services = append(services, model.NewService(proto))
		}
	}
	return services
}

func (d *FileDescriptorSet) GetMessages() model.Messages {
	var messages model.Messages
	for _, f := range d.GetFile() {
		for _, proto := range f.GetMessageType() {
			messages = append(messages, model.NewMessage(proto))
		}
	}
	return messages
}
