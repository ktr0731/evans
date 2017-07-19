package repl

import (
	"github.com/lycoris0731/evans/lib/model"
	"github.com/lycoris0731/evans/lib/parser"
	"github.com/pkg/errors"
)

var (
	ErrUnselected = errors.New("unselected")
)

// packages is used by showing all packages
// mapPackages is used by extract a package by package name
type cache struct {
	packages    model.Packages
	mapPackages map[string]model.Package
}

type Env struct {
	Desc  *parser.FileDescriptorSet
	state state

	cache cache
}

func (e *Env) GetPackages() model.Packages {
	if e.cache.packages != nil {
		return e.cache.packages
	}

	packNames := e.Desc.GetPackages()
	packages := make(model.Packages, len(packNames))
	for i, name := range packNames {
		packages[i] = &model.Package{Name: name}
	}

	e.cache.packages = packages

	return packages
}

func (e *Env) GetServices() (model.Services, error) {
	if e.state.currentPackage == "" {
		return nil, errors.Wrap(ErrUnselected, "package")
	}

	name := e.state.currentPackage

	pack, ok := e.cache.mapPackages[name]
	if ok {
		return pack.Services, nil
	}

	pack.Services = e.Desc.GetServices(name)

	return e.Desc.GetServices(e.state.currentPackage), nil
}

// func (e *Env) GetMessages() (model.Messages, error) {
// 	if e.state.currentPackage == "" {
// 		return nil, errors.Wrap(ErrUnselected, "package")
// 	}
//
// 	name := e.state.currentPackage
//
// 	pack, ok := e.cache.mapPackages[name]
// 	if ok {
// 		return pack.Services, nil
// 	}
//
// 	pack.Messages = e.Desc.GetMessages(name)
//
// 	return e.Desc.GetServices(e.state.currentPackage), nil
// }
