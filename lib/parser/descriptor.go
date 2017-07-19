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

func (d *FileDescriptorSet) GetService(packName string, svcName string) (*model.Service, error) {
	// for _, svc := range d.GetServices() {
	// 	if svc.Name == name {
	// 		return svc, nil
	// 	}
	// }
	// return nil, fmt.Errorf("service %s not found", name)
	return nil, nil
}
func (d *FileDescriptorSet) GetRPC(svcName, rpcName string) (*model.RPC, error) {
	// services, err := d.GetService(svcName)
	// if err != nil {
	// 	return nil, err
	// }
	// for _, rpc := range services.RPCs {
	// 	if rpc.Name == rpcName {
	// 		return &rpc, nil
	// 	}
	// }
	// return nil, fmt.Errorf("RPC %s in service %s not found", rpcName, svcName)
	return nil, nil
}

func (d *FileDescriptorSet) GetMessage(pack, name string) *model.Message {
	// TODO: 自前 descriptor の GetMessage 引くようにする
	return model.NewMessage(d.FileDescriptorSet.GetMessage(pack, name))
}
