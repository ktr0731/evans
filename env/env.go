package env

import (
	"strings"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/parser"
	"github.com/pkg/errors"
)

var (
	ErrPackageUnselected    = errors.New("package unselected")
	ErrServiceUnselected    = errors.New("service unselected")
	ErrUnknownTarget        = errors.New("unknown target")
	ErrUnknownPackage       = errors.New("unknown package")
	ErrUnknownService       = errors.New("unknown service")
	ErrInvalidServiceName   = errors.New("invalid service name")
	ErrInvalidMessageName   = errors.New("invalid message name")
	ErrInvalidRPCName       = errors.New("invalid RPC name")
	ErrServiceCachingFailed = errors.New("service caching failed")
)

// pkgList is used by showing all packages
// pkg is used by extract a package by package name
type cache struct {
	pkgList entity.Packages
	pkg     map[string]*entity.Package
}

type state struct {
	currentPackage string
	currentService string
}

type Env struct {
	desc  *parser.FileDescriptorSet
	state state

	config *config.Env

	cache cache
}

func New(desc *parser.FileDescriptorSet, config *config.Env) (*Env, error) {
	return &Env{
		desc:   desc,
		config: config,
		cache: cache{
			pkg: map[string]*entity.Package{},
		},
	}, nil
}

func (e *Env) HasCurrentPackage() bool {
	return e.state.currentPackage != ""
}

func (e *Env) HasCurrentService() bool {
	return e.state.currentService != ""
}

func (e *Env) GetPackages() entity.Packages {
	if e.cache.pkgList != nil {
		return e.cache.pkgList
	}

	pkgNames := e.desc.GetPackages()
	pkgs := make(entity.Packages, len(pkgNames))
	for i, name := range pkgNames {
		pkgs[i] = &entity.Package{Name: name}
	}

	e.cache.pkgList = pkgs

	return pkgs
}

func (e *Env) GetServices() (entity.Services, error) {
	if !e.HasCurrentPackage() {
		return nil, ErrPackageUnselected
	}

	// services, messages and rpc are cached to e.cache when called UsePackage()
	// if messages isn't cached, it occurred panic
	return e.cache.pkg[e.state.currentPackage].Services, nil
}

func (e *Env) GetMessages() (entity.Messages, error) {
	// TODO: current package 以外からも取得したい
	if !e.HasCurrentPackage() {
		return nil, ErrPackageUnselected
	}

	// same as GetServices()
	return e.cache.pkg[e.state.currentPackage].Messages, nil
}

func (e *Env) GetRPCs() (entity.RPCs, error) {
	if !e.HasCurrentService() {
		return nil, ErrServiceUnselected
	}

	svc, err := e.GetService(e.state.currentService)
	if err != nil {
		return nil, err
	}
	return svc.RPCs, nil
}

func (e *Env) GetService(name string) (*entity.Service, error) {
	svc, err := e.GetServices()
	if err != nil {
		return nil, err
	}
	for _, svc := range svc {
		if name == svc.Name {
			return svc, nil
		}
	}
	return nil, errors.Wrapf(ErrInvalidServiceName, "%s not found", name)
}

func (e *Env) GetMessage(name string) (*entity.Message, error) {
	// Person2 で panic
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
	return nil, errors.Wrapf(ErrInvalidMessageName, "%s not found", name)
}

func (e *Env) GetRPC(name string) (*entity.RPC, error) {
	rpcs, err := e.GetRPCs()
	if err != nil {
		return nil, err
	}
	for _, rpc := range rpcs {
		if name == rpc.Name {
			return rpc, nil
		}
	}
	return nil, errors.Wrapf(ErrInvalidRPCName, "%s not found", name)
}

func (e *Env) UsePackage(name string) error {
	for _, p := range e.desc.GetPackages() {
		if name == p {
			e.state.currentPackage = name
			return e.loadPackage(p)
		}
	}
	return errors.Wrapf(ErrUnknownPackage, "%s not found", name)
}

func (e *Env) UseService(name string) error {
	// set extracted package if passed service which has package name
	if e.state.currentPackage == "" {
		s := strings.SplitN(name, ".", 2)
		if len(s) != 2 {
			return errors.Wrap(ErrPackageUnselected, "please set package (package_name.service_name or set --package flag)")
		}
		if err := e.UsePackage(s[0]); err != nil {
			return errors.Wrapf(err, name)
		}
	}
	services, err := e.GetServices()
	if err != nil {
		return errors.Wrapf(err, "failed to get services")
	}
	for _, svc := range services {
		if name == svc.Name {
			e.state.currentService = name
			return nil
		}
	}
	return errors.Wrapf(ErrUnknownService, "%s not found", name)
}

func (e *Env) GetDSN() string {
	if e.state.currentPackage == "" {
		return ""
	}
	dsn := e.state.currentPackage
	if e.state.currentService != "" {
		dsn += "." + e.state.currentService
	}
	return dsn
}

// loadPackage loads all services and messages in itself
func (e *Env) loadPackage(name string) error {
	// prevent duplicated loading
	_, ok := e.cache.pkg[name]
	if ok {
		return nil
	}

	dSvc := e.desc.GetServices(name)
	dMsg := e.desc.GetMessages(name)

	// Messages: actual message size is greater than or equal to len(dMsg)
	//           because message can be contain other messages as a field
	e.cache.pkg[name] = &entity.Package{
		Name:     name,
		Services: make(entity.Services, len(dSvc)),
		Messages: make(entity.Messages, 0, len(dMsg)),
	}

	services := make(entity.Services, len(dSvc))
	for i, svc := range dSvc {
		services[i] = entity.NewService(svc)
		services[i].RPCs = entity.NewRPCs(svc)
	}
	e.cache.pkg[name].Services = services

	messages := make(entity.Messages, len(dMsg))
	for i, msg := range dMsg {
		messages[i] = entity.NewMessage(msg)

		fields, err := entity.NewFields(e.cache.pkg[name], messages[i])
		if err != nil {
			return errors.Wrapf(err, "failed to get field of %s", msg.GetName())
		}
		messages[i].Fields = fields

		// cache each result by each time because some messages depends on some messages
		e.cache.pkg[name].Messages = append(e.cache.pkg[name].Messages, messages[i])
	}

	return nil
}

// Full Qualified Name
// It contains message or service with package name
// e.g.: .test.Person -> Person
func (e *Env) getNameFromFQN(fqn string) string {
	return strings.TrimLeft(fqn, "."+e.state.currentPackage+".")
}

// getMessage is a closure which has current states
// it is passed by entity.NewField() for get message from current package
func (e *Env) getMessage() func(typeName string) (*entity.Message, error) {
	return func(msgName string) (*entity.Message, error) {
		return e.GetMessage(msgName)
	}
}

func (e *Env) getService() func(typeName string) (*entity.Service, error) {
	return func(svcName string) (*entity.Service, error) {
		return e.GetService(svcName)
	}
}
