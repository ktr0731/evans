package port

import (
	"io"

	"github.com/ktr0731/evans/entity"
)

type InputPort interface {
	Package(*PackageParams) (io.Reader, error)
	Service(*ServiceParams) (io.Reader, error)

	Describe(*DescribeParams) (io.Reader, error)
	Show(*ShowParams) (io.Reader, error)

	Header(*HeaderParams) (io.Reader, error)

	Call(*CallParams) (io.Reader, error)
}

type CallParams struct {
	RPCName string
}

type DescribeParams struct {
	MsgName string
}

type PackageParams struct {
	PkgName string
}

type ServiceParams struct {
	SvcName string
}

type ShowType int

const (
	ShowTypePackage = iota
	ShowTypeService
	ShowTypeMessage
	ShowTypeRPC
	ShowTypeHeader
)

type Showable interface {
	Show() string
}

type ShowParams struct {
	Type ShowType
}

type HeaderParams struct {
	Headers []*entity.Header
}
