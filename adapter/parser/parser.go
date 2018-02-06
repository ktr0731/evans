package parser

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/adapter/internal/proto_parser"
	"github.com/ktr0731/evans/entity"
)

func ParseFile(filename []string, paths []string) (entity.Packages, error) {
	set, err := proto_parser.ParseFile(filename, paths)
	if err != nil {
		return nil, err
	}
	return toEntitiesFrom(set)
}

// toEntitiesFrom normalizes descriptors to entities
//
// package
// ├ messages
// ├ enums
// └ services
//   └ rpcs
//
func toEntitiesFrom(files []*desc.FileDescriptor) (entity.Packages, error) {
	var pkgNames []string
	msgMap := map[string][]*entity.Message{}
	svcMap := map[string][]*entity.Service{}
	for _, f := range files {
		pkgName := f.GetPackage()

		pkgNames = append(pkgNames, pkgName)

		for _, msg := range f.GetMessageTypes() {
			msgMap[pkgName] = append(msgMap[pkgName], newMessage(msg))
		}
		for _, svc := range f.GetServices() {
			svcMap[pkgName] = append(svcMap[pkgName], newService(svc))
		}
	}

	var pkgs Packages
	for _, pkgName := range pkgNames {
		pkgs = append(pkgs, newPackage(pkgName, msgMap[pkgName], svcMap[pkgName]))
	}

	return pkgs, nil
}
