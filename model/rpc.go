package model

import (
	"bytes"

	"github.com/jhump/protoreflect/desc"
	"github.com/olekukonko/tablewriter"
)

type RPC struct {
	Name         string
	RequestType  *desc.MessageDescriptor
	ResponseType *desc.MessageDescriptor
}

type RPCs []*RPC

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
