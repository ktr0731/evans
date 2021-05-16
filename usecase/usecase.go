// Package usecase provides all use-case logics used from each mode.
// Clients of usecase package must call Inject before all function calls in usecase package.
package usecase

import (
	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/format"
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
	interactiveFiller fill.InteractiveFiller
	gRPCClient        grpc.Client
	responseFormatter *format.ResponseFormatter
	resourcePresenter present.Presenter
	state             state
}

type rpcIdentifier string

// state has the domain state modified by each usecase logic. The default value is used as the initial value.
type state struct {
	selectedPackage string // TODO: remove in v1.0.0.
	selectedService string
	rpcCallState    map[rpcIdentifier]callState
}

type callState struct {
	reqPayload []byte
	repeatable bool
}

type Dependencies struct {
	Spec              idl.Spec
	Filler            fill.Filler
	InteractiveFiller fill.InteractiveFiller
	GRPCClient        grpc.Client
	ResponseFormatter *format.ResponseFormatter
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
		interactiveFiller: d.InteractiveFiller,
		gRPCClient:        d.GRPCClient,
		responseFormatter: d.ResponseFormatter,
		resourcePresenter: d.ResourcePresenter,

		state: defaultState,
	}
}

// InjectPartially is almost same as the Inject, but injects only non-nil dependencies.
func InjectPartially(deps Dependencies) {
	dm.InjectPartially(deps)
}

func (m *dependencyManager) InjectPartially(d Dependencies) {
	if d.Spec != nil {
		m.spec = d.Spec
	}
	if d.Filler != nil {
		m.filler = d.Filler
	}
	if d.InteractiveFiller != nil {
		m.interactiveFiller = d.InteractiveFiller
	}
	if d.GRPCClient != nil {
		m.gRPCClient = d.GRPCClient
	}
	if d.ResponseFormatter != nil {
		m.responseFormatter = d.ResponseFormatter
	}
	if d.ResourcePresenter != nil {
		m.resourcePresenter = d.ResourcePresenter
	}
}

// Clear clears all dependencies and states. Usually, it is used for unit testing.
func Clear() {
	dm.Inject(Dependencies{})
}
