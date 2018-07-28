package entity

import (
	"sort"
	"strings"
	"sync"

	"github.com/ktr0731/evans/config"
	"github.com/pkg/errors"
)

var (
	ErrPackageUnselected  = errors.New("package unselected")
	ErrServiceUnselected  = errors.New("service unselected")
	ErrUnknownPackage     = errors.New("unknown package")
	ErrUnknownService     = errors.New("unknown service")
	ErrInvalidServiceName = errors.New("invalid service name")
	ErrInvalidMessageName = errors.New("invalid message name")
	ErrInvalidRPCName     = errors.New("invalid RPC name")
)

type Environment interface {
	Packages() []*Package
	Services() ([]Service, error)
	Messages() ([]Message, error)
	RPCs() ([]RPC, error)
	Service(name string) (Service, error)
	Message(name string) (Message, error)
	RPC(name string) (RPC, error)

	Headers() []*Header
	AddHeader(header *Header)
	RemoveHeader(key string)

	UsePackage(name string) error
	UseService(name string) error

	DSN() string
}

// pkgList is used by showing all packages
// pkg is used by extract a package by package name
type cache struct {
	pkgList []*Package
	pkg     map[string]*Package
}

type state struct {
	currentPackage string
	currentService string
}

type option struct {
	headers sync.Map
}

type Env struct {
	pkgs   []*Package
	state  state
	option option
	config *config.Env
	cache  cache
}

func NewEnv(pkgs []*Package, config *config.Config) *Env {
	env := &Env{
		pkgs:   pkgs,
		config: config.Env,
		cache: cache{
			pkg: map[string]*Package{},
		},
	}

	for _, h := range config.Request.Header {
		env.AddHeader(&Header{Key: h.Key, Val: h.Val})
	}

	return env
}

// NewEnvFromServices is called if the target server has enabled gRPC reflection.
// gRPC reflection has no packages, so Evans creates pseudo package "default".
func NewEnvFromServices(svcs []Service, config *config.Config) *Env {
	env := NewEnv([]*Package{
		{
			Name:     "default",
			Services: svcs,
		},
	}, config)

	err := env.UsePackage(env.pkgs[0].Name)
	if err != nil {
		panic(err)
	}

	return env
}

func (e *Env) HasCurrentPackage() bool {
	return e.state.currentPackage != ""
}

func (e *Env) HasCurrentService() bool {
	return e.state.currentService != ""
}

func (e *Env) Packages() []*Package {
	return e.pkgs
}

func (e *Env) Services() ([]Service, error) {
	if !e.HasCurrentPackage() {
		return nil, ErrPackageUnselected
	}

	// services, messages and rpc are cached to e.cache when called UsePackage()
	// if messages isn't cached, it occurred panic
	return e.cache.pkg[e.state.currentPackage].Services, nil
}

func (e *Env) Messages() ([]Message, error) {
	if !e.HasCurrentPackage() {
		return nil, ErrPackageUnselected
	}

	return e.cache.pkg[e.state.currentPackage].Messages, nil
}

func (e *Env) RPCs() ([]RPC, error) {
	if !e.HasCurrentService() {
		return nil, ErrServiceUnselected
	}

	svc, err := e.Service(e.state.currentService)
	if err != nil {
		return nil, err
	}
	return svc.RPCs(), nil
}

func (e *Env) Service(name string) (Service, error) {
	svc, err := e.Services()
	if err != nil {
		return nil, err
	}
	for _, svc := range svc {
		if name == svc.Name() {
			return svc, nil
		}
	}
	return nil, errors.Wrapf(ErrInvalidServiceName, "%s not found", name)
}

func (e *Env) Message(name string) (Message, error) {
	msg, err := e.Messages()
	if err != nil {
		return nil, err
	}
	for _, msg := range msg {
		if name == msg.Name() {
			return msg, nil
		}
	}
	return nil, errors.Wrapf(ErrInvalidMessageName, "%s not found", name)
}

func (e *Env) Headers() (headers []*Header) {
	e.option.headers.Range(func(k, v interface{}) bool {
		h := v.(*Header)
		headers = append(headers, &Header{Key: h.Key, Val: h.Val})
		return true
	})
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].Key < headers[j].Key
	})
	return
}

func (e *Env) AddHeader(h *Header) {
	e.option.headers.Store(h.Key, h)
}

func (e *Env) RemoveHeader(key string) {
	e.option.headers.Delete(key)
}

func (e *Env) RPC(name string) (RPC, error) {
	rpcs, err := e.RPCs()
	if err != nil {
		return nil, err
	}
	for _, rpc := range rpcs {
		if name == rpc.Name() {
			return rpc, nil
		}
	}
	return nil, errors.Wrapf(ErrInvalidRPCName, "%s not found", name)
}

func (e *Env) UsePackage(name string) error {
	for _, p := range e.Packages() {
		if name == p.Name {
			e.state.currentPackage = name
			e.cache.pkg[name] = p
			return nil
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
	services, err := e.Services()
	if err != nil {
		return errors.Wrapf(err, "failed to get services")
	}
	for _, svc := range services {
		if name == svc.Name() {
			e.state.currentService = name
			return nil
		}
	}
	return errors.Wrapf(ErrUnknownService, "%s not found", name)
}

func (e *Env) DSN() string {
	if e.state.currentPackage == "" {
		return ""
	}
	dsn := e.state.currentPackage
	if e.state.currentService != "" {
		dsn += "." + e.state.currentService
	}
	return dsn
}
