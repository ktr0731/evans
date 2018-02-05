package usecase

import (
	"bytes"
	"errors"
	"io"

	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/olekukonko/tablewriter"
)

func Show(params *port.ShowParams, outputPort port.OutputPort, env entity.Environment) (io.Reader, error) {
	var showable port.Showable
	switch params.Type {
	case port.ShowTypePackage:
		showable = packages(env.Packages())
	case port.ShowTypeService:
		svcs, err := env.Services()
		if err != nil {
			return nil, err
		}
		showable = services(svcs)
	case port.ShowTypeMessage:
		msgs, err := env.Messages()
		if err != nil {
			return nil, err
		}
		showable = messages(msgs)
	case port.ShowTypeRPC:
		r, err := env.RPCs()
		if err != nil {
			return nil, err
		}
		showable = rpcs(r)
	case port.ShowTypeHeader:
		showable = headers(env.Headers())
	default:
		return nil, errors.New("unknown showable type")
	}

	return outputPort.Show(showable)
}

type packages entity.Packages

func (p packages) Show() string {
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

type services entity.Services

func (s services) Show() string {
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

type message struct {
	*entity.Message
}

func (m *message) Show() string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"field", "type"})
	rows := [][]string{}
	for _, f := range m.Fields() {
		row := []string{f.Name, f.Type}
		rows = append(rows, row)
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}

type messages entity.Messages

func (m messages) Show() string {
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

type rpcs entity.RPCs

func (r rpcs) Show() string {
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

type headers []*entity.Header

func (h headers) Show() string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"key", "val"})
	rows := [][]string{}
	for _, header := range h {
		row := []string{header.Key, header.Val}
		rows = append(rows, row)
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}
