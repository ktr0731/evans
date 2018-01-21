package entity

import (
	"github.com/jhump/protoreflect/desc"
)

type Parser interface {
	ParseFile(fnames []string, fpaths []string) (Packages, error)
}

// ToEntitiesFrom normalizes descriptors to entities
//
// package
// ├ messages
// ├ enums
// └ services
//   └ rpcs
//
func ToEntitiesFrom(files []*desc.FileDescriptor) (Packages, error) {
	var pkgNames []string
	msgMap := map[string][]*Message{}
	svcMap := map[string][]*Service{}
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
		pkgs = append(pkgs, NewPackage(pkgName, msgMap[pkgName], svcMap[pkgName]))
	}

	return pkgs, nil
}
