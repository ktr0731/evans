package entity

import (
	"bytes"

	"github.com/jhump/protoreflect/desc"
	"github.com/olekukonko/tablewriter"
)

type RPC struct {
	Name         string
	FQRN         string
	RequestType  *desc.MessageDescriptor
	ResponseType *desc.MessageDescriptor
}

type RPCs []*RPC

// NewRPCs collects RPCs in ServiceDescriptor.
// Only NewRPCs receive a raw descriptor because it is called by NewService.
// So, NewRPCs needs to receive a raw descriptor instead of entity.Service.
func NewRPCs(svc *desc.ServiceDescriptor) RPCs {
	rpcs := make(RPCs, len(svc.GetMethods()))
	for i, rpc := range svc.GetMethods() {
		r := &RPC{
			Name:         rpc.GetName(),
			FQRN:         rpc.GetFullyQualifiedName(),
			RequestType:  rpc.GetInputType(),
			ResponseType: rpc.GetOutputType(),
		}
		rpcs[i] = r
	}
	return rpcs
}

func (r *RPCs) String() string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"RPC", "RequestType", "ResponseType"})
	rows := [][]string{}
	for _, rpc := range *r {
		row := []string{rpc.Name, rpc.RequestType.GetName(), rpc.ResponseType.GetName()}
		rows = append(rows, row)
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}
