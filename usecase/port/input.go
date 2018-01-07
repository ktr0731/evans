package port

import (
	"fmt"
	"io"

	"github.com/ktr0731/evans/entity"
)

type InputPort interface {
	Package(*PackageParams) (*PackageResponse, error)
	Service(*ServiceParams) (*ServiceResponse, error)

	Describe(*DescribeParams) (*DescribeResponse, error)
	Show(*ShowParams) (*ShowResponse, error)

	Header(*HeaderParams) (*HeaderResponse, error)

	Call(*CallParams) (io.Reader, error)
}

type CallParams struct {
	RPCName string
}

type CallResponse struct{}

type DescribeParams struct {
	MsgName string
}

type DescribeResponse struct{}

type PackageParams struct {
	PkgName string
}

type PackageResponse struct{}

type ServiceParams struct {
	SvcName string
}

type ServiceResponse struct{}

type ShowType int

const (
	ShowTypePackage = iota
	ShowTypeService
	ShowTypeMessage
	ShowTypeRPC
)

type Showable interface {
	canShow() bool // used as identifier only
	fmt.Stringer
}

type Packages entity.Packages

func (p Packages) canShow() bool {
	return true
}

func (p Packages) String() string {
	panic("not implemented yet")
	return ""
}

type Services entity.Services

func (s Services) canShow() bool {
	return true
}

func (s Services) String() string {
	panic("not implemented yet")
	return ""
}

type Messages entity.Messages

func (m Messages) canShow() bool {
	return true
}

func (m Messages) String() string {
	panic("not implemented yet")
	return ""
}

type RPCs entity.RPCs

func (r RPCs) canShow() bool {
	return true
}

func (r RPCs) String() string {
	panic("not implemented yet")
	return ""
}

type ShowParams struct {
	Name string
	Type ShowType
}

type ShowResponse struct{}

type HeaderParams struct {
	Key string
	Val string
}

type HeaderResponse struct{}
