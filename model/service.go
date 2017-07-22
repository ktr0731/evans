package model

import (
	"bytes"

	"github.com/jhump/protoreflect/desc"
	"github.com/olekukonko/tablewriter"
)

type Service struct {
	Name string
	RPCs RPCs

	Desc *desc.ServiceDescriptor
}

func NewService(service *desc.ServiceDescriptor) *Service {
	return &Service{
		Name: service.GetName(),
		RPCs: NewRPCs(service),
	}
}

type Services []*Service

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
