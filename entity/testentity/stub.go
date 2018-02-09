package testentity

import (
	"math/rand"

	"github.com/ktr0731/evans/entity"
)

var names []string = []string{
	"ayatsuji",
	"sakurai",
	"tanamachi",
	"nakata",
	"nanasaki",
	"morishima",
}

func newName() string {
	return names[rand.Intn(len(names))]
}

type Fld struct {
	entity.Field

	FName       string
	FFieldName  string
	FPBtype     string
	FIsRepeated bool
}

func NewFld() *Fld {
	return &Fld{
		FName:       newName(),
		FFieldName:  newName(),
		FPBtype:     newName(),
		FIsRepeated: false,
	}
}

func (f *Fld) Name() string {
	return f.FName
}

func (f *Fld) FieldName() string {
	return f.FFieldName
}

func (f *Fld) PBType() string {
	return f.FPBtype
}

func (f *Fld) IsRepeated() bool {
	return f.FIsRepeated
}

type RPC struct {
	entity.RPC

	name     string
	req, res entity.Message
	fqrn     string
}

func NewRPC() *RPC {
	return &RPC{
		name: newName(),
		req:  NewMsg(),
		res:  NewMsg(),
		fqrn: newName(),
	}
}

func (r *RPC) Name() string {
	return r.name
}

func (r *RPC) RequestMessage() entity.Message {
	return r.req
}

func (r *RPC) ResponseMessage() entity.Message {
	return r.res
}

func (r *RPC) FQRN() string {
	return r.fqrn
}

type Svc struct {
	entity.Service

	name string
	rpcs []*RPC
}

func NewSvc() *Svc {
	rpcs := NewRPCs()
	return &Svc{
		name: newName(),
		rpcs: rpcs,
	}
}

func (s *Svc) Name() string {
	return s.name
}

func (s *Svc) RPCs() []entity.RPC {
	rpcs := make([]entity.RPC, 0, len(s.rpcs))
	for _, rpc := range s.rpcs {
		rpcs = append(rpcs, rpc)
	}
	return rpcs
}

func NewRPCs() []*RPC {
	rpcs := make([]*RPC, 0, rand.Intn(3))
	for i := 0; i < len(rpcs); i++ {
		rpcs = append(rpcs, NewRPC())
	}
	return rpcs
}

type Msg struct {
	entity.Message

	name   string
	fields []entity.Field
}

func NewMsg() *Msg {
	flds := make([]entity.Field, 0, rand.Intn(3))
	for i := 0; i < len(flds); i++ {
		flds = append(flds, NewFld())
	}
	return &Msg{name: newName(), fields: flds}
}

func (m *Msg) Name() string {
	return m.name
}

func (m *Msg) Fields() []entity.Field {
	return m.fields
}
