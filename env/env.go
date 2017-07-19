package env

import (
	"log"

	"github.com/lycoris0731/evans/lib/model"
	"github.com/lycoris0731/evans/lib/parser"
	"github.com/pkg/errors"
)

var (
	ErrUnselected     = errors.New("unselected")
	ErrUnknownTarget  = errors.New("unknown target")
	ErrUnknownPackage = errors.New("unknown package")
	ErrUnknownService = errors.New("unknown service")
)

// packages is used by showing all packages
// mapPackages is used by extract a package by package name
type cache struct {
	packages    model.Packages
	mapPackages map[string]*model.Package
}

type State struct {
	currentPackage string
	currentService string
}

type Env struct {
	Desc *parser.FileDescriptorSet
	*State

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
	if e.currentPackage == "" {
		return nil, errors.Wrap(ErrUnselected, "package")
	}

	name := e.currentPackage

	pack, ok := e.cache.mapPackages[name]
	if ok {
		log.Println("use cache")
		return pack.Services, nil
	}

	log.Println("no cache")
	pack.Services = e.Desc.GetServices(name)

	return pack.Services, nil
}

func (e *Env) GetMessages() (model.Messages, error) {
	if e.currentPackage == "" {
		return nil, errors.Wrap(ErrUnselected, "package")
	}

	name := e.currentPackage

	pack, ok := e.cache.mapPackages[name]
	if ok {
		return pack.Messages, nil
	}

	pack.Messages = e.Desc.GetMessages(name)

	return pack.Messages, nil
}

func (e *Env) UsePackage(name string) error {
	for _, p := range e.Desc.GetPackages() {
		if name == p {
			e.currentPackage = name
			return nil
		}
	}
	return ErrUnknownPackage
}

func (e *Env) UseService(name string) error {
	// for _, svc := range r.env.Desc.GetServices() {
	// 	if name == svc.Name {
	// 		r.env.state.currentService = name
	// 		return nil
	// 	}
	// }
	return ErrUnknownService
}

func (e *Env) GetDSN() string {
	if e.currentPackage == "" {
		return ""
	}
	dsn := e.currentPackage
	if e.currentService != "" {
		dsn += "." + e.currentService
	}
	return dsn
}
