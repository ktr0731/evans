package model

import (
	"bytes"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/olekukonko/tablewriter"
)

type RPC struct {
	Name         string
	RequestType  string
	ResponseType string
}

type Service struct {
	Name string
	RPCs []RPC
}

func NewService(service *descriptor.ServiceDescriptorProto) *Service {
	rpcs := make([]RPC, len(service.GetMethod()))
	for i, rpc := range service.GetMethod() {
		rpcs[i] = RPC{
			Name:         rpc.GetName(),
			RequestType:  rpc.GetInputType(),
			ResponseType: rpc.GetOutputType(),
		}
	}
	return &Service{
		Name: service.GetName(),
		RPCs: rpcs,
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
			row := []string{serviceName, rpc.Name, rpc.RequestType, rpc.ResponseType}
			rows = append(rows, row)
		}
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}
