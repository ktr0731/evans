// Package usecase provides all use-case logics used from each mode.
// Clients of usecase package must call Inject before all function calls in usecase package.
package usecase

import (
	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/grpc"
	"github.com/ktr0731/evans/idl"
	"github.com/ktr0731/evans/present"
)

var (
	defaultState = state{
		headers: grpc.Headers{},
	}
	dm = &dependencyManager{}
)

type dependencyManager struct {
	spec       idl.Spec
	filler     fill.Filler
	gRPCClient grpc.Client
	presenter  present.Presenter

	state state
}

// state has the domain state modified by each usecase logic. The default value is used as the initial value.
type state struct {
	selectedPackage string
	selectedService string
	headers         grpc.Headers
}

// clone copies itself deeply.
func (s state) clone() state {
	headers := grpc.Headers{}
	for k, v := range s.headers {
		headers[k] = v
	}
	return state{
		headers: headers,
	}
}

// Inject corresponds an implementation to an interface type. Inject clears the previous states if it exists.
func Inject(
	spec idl.Spec,
	filler fill.Filler,
	gRPCClient grpc.Client,
	presenter present.Presenter,
) {
	dm.Inject(spec, filler, gRPCClient, presenter)
}

func (m *dependencyManager) Inject(
	spec idl.Spec,
	filler fill.Filler,
	gRPCClient grpc.Client,
	presenter present.Presenter,
) {
	dm = &dependencyManager{
		spec:       spec,
		filler:     filler,
		gRPCClient: gRPCClient,
		presenter:  presenter,

		state: defaultState.clone(),
	}
}

// Clear clears all dependencies and states. Usually, it is used for unit testing.
func Clear() {
	dm.Inject(nil, nil, nil, nil)
}
