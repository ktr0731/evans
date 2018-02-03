package port

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ktr0731/evans/entity"
	"github.com/olekukonko/tablewriter"
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
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"package"})
	rows := [][]string{}
	for _, pack := range p {
		row := []string{pack.Name}
		rows = append(rows, row)
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}

type Services entity.Services

func (s Services) canShow() bool {
	return true
}

func (s Services) String() string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"service", "RPC", "RequestType", "ResponseType"})
	rows := [][]string{}
	for _, service := range s {
		first := true
		for _, rpc := range service.RPCs {
			serviceName := ""
			if first {
				serviceName = service.Name
				first = false
			}
			row := []string{serviceName, rpc.Name, rpc.RequestType.GetName(), rpc.ResponseType.GetName()}
			rows = append(rows, row)
		}
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}

type Messages entity.Messages

func (m Messages) canShow() bool {
	return true
}

func (m Messages) String() string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"message"})
	rows := [][]string{}
	for _, message := range m {
		row := []string{message.Name()}
		rows = append(rows, row)
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}

type RPCs entity.RPCs

func (r RPCs) canShow() bool {
	return true
}

func (r RPCs) String() string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"RPC", "RequestType", "ResponseType"})
	rows := [][]string{}
	for _, rpc := range r {
		row := []string{rpc.Name, rpc.RequestType.GetName(), rpc.ResponseType.GetName()}
		rows = append(rows, row)
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}

type ShowParams struct {
	Type ShowType
}

type HeaderParams struct {
	Headers []*entity.Header
}
