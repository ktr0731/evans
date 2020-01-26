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
	defaultState = state{}
	dm           = &dependencyManager{}
)

type dependencyManager struct {
	spec              idl.Spec
	filler            fill.Filler
	gRPCClient        grpc.Client
	responsePresenter present.Presenter
	resourcePresenter present.Presenter

	state state
}

// state has the domain state modified by each usecase logic. The default value is used as the initial value.
type state struct {
	selectedPackage string
	selectedService string
}

type Dependencies struct {
	Spec              idl.Spec
	Filler            fill.Filler
	GRPCClient        grpc.Client
	ResponsePresenter present.Presenter
	ResourcePresenter present.Presenter
}

// Inject corresponds an implementation to an interface type. Inject clears the previous states if it exists.
func Inject(deps Dependencies) {
	dm.Inject(deps)
}

func (m *dependencyManager) Inject(d Dependencies) {
	dm = &dependencyManager{
		spec:              d.Spec,
		filler:            d.Filler,
		gRPCClient:        d.GRPCClient,
		responsePresenter: d.ResponsePresenter,
		resourcePresenter: d.ResourcePresenter,

		state: defaultState,
	}
}

// Clear clears all dependencies and states. Usually, it is used for unit testing.
func Clear() {
	dm.Inject(Dependencies{})
}
