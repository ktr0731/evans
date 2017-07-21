package env

import (
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/lycoris0731/evans/lib/parser"
	"github.com/lycoris0731/evans/model"
	"github.com/pkg/errors"
)

var (
	ErrUnselected         = errors.New("unselected")
	ErrUnknownTarget      = errors.New("unknown target")
	ErrUnknownPackage     = errors.New("unknown package")
	ErrUnknownService     = errors.New("unknown service")
	ErrInvalidServiceName = errors.New("invalid service name")
	ErrInvalidMessageName = errors.New("invalid message name")
	ErrInvalidRPCName     = errors.New("invalid RPC name")
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
	desc *parser.FileDescriptorSet
	State

	cache cache
}

func NewEnv(desc *parser.FileDescriptorSet) *Env {
	env := &Env{desc: desc}
	env.cache.mapPackages = map[string]*model.Package{}
	return env
}

func (e *Env) GetPackages() model.Packages {
	if e.cache.packages != nil {
		return e.cache.packages
	}

	packNames := e.desc.GetPackages()
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

	// services, messages and rpc are cached by Env on startup
	name := e.currentPackage
	pkg, ok := e.cache.mapPackages[name]
	if ok {
		return pkg.Services, nil
	}

	return nil, errors.New("caching failed")
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

	return nil, errors.New("caching failed")
}

func (e *Env) GetRPCs() (model.RPCs, error) {
	if e.currentService == "" {
		return nil, errors.Wrap(ErrUnselected, "service")
	}

	name := e.currentService
	svc, err := e.GetService(name)
	if err != nil {
		return nil, err
	}
	return svc.RPCs, nil
}

func (e *Env) GetService(name string) (*model.Service, error) {
	svc, err := e.GetServices()
	if err != nil {
		return nil, err
	}
	for _, svc := range svc {
		if name == svc.Name {
			return svc, nil
		}
	}
	return nil, errors.Wrap(ErrInvalidServiceName, name)
}

func (e *Env) GetMessage(name string) (*model.Message, error) {
	msg, err := e.GetMessages()
	if err != nil {
		return nil, err
	}
	for _, msg := range msg {
		msgName := e.getNameFromFQN(name)
		if msgName == msg.Name {
			return msg, nil
		}
	}
	return nil, errors.Wrap(ErrInvalidMessageName, name)
}

func (e *Env) GetRPC(name string) (*model.RPC, error) {
	rpcs, err := e.GetRPCs()
	if err != nil {
		return nil, err
	}
	for _, rpc := range rpcs {
		if name == rpc.Name {
			return rpc, nil
		}
	}
	return nil, errors.Wrap(ErrInvalidRPCName, name)
}

func (e *Env) UsePackage(name string) error {
	for _, p := range e.desc.GetPackages() {
		if name == p {
			e.currentPackage = name
			return e.loadPackage(p)
		}
	}
	return ErrUnknownPackage
}

func (e *Env) UseService(name string) error {
	for _, svc := range e.desc.GetServices(e.currentPackage) {
		if name == svc.GetName() {
			e.currentService = name
			return nil
		}
	}
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

// loadPackage loads all services and messages in itself
func (e *Env) loadPackage(name string) error {
	dSvc := e.desc.GetServices(name)
	dMsg := e.desc.GetMessages(name)

	services := make(model.Services, len(dSvc))
	for i, svc := range dSvc {
		services[i] = model.NewService(svc)
	}

	messages := make(model.Messages, len(dMsg))
	for i, msg := range dMsg {
		fmt.Println(msg.GetName())
		messages[i] = model.NewMessage(msg)
		messages[i].Fields = model.NewFields(e.getMessage(e.currentPackage), msg)
	}

	_, ok := e.cache.mapPackages[name]
	if ok {
		return errors.New("duplicated loading")
	}
	e.cache.mapPackages[name] = &model.Package{
		Name:     name,
		Services: services,
		Messages: messages,
	}

	return nil
}

// Full Qualified Name
// It contains message or service with package name
// e.g.: .test.Person
func (e *Env) getNameFromFQN(fqn string) string {
	return strings.TrimLeft(fqn, "."+e.currentPackage+".")
}

func (e *Env) getMessage(pkgName string) func(typeName string) *desc.MessageDescriptor {
	messages := e.desc.GetMessages(pkgName)

	return func(msgName string) *desc.MessageDescriptor {
		for _, msg := range messages {
			if msg.GetName() == msgName {
				return msg
			}
		}
		return nil
	}
}
