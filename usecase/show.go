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
		showable = port.Packages(env.Packages())
	case port.ShowTypeService:
		svcs, err := env.Services()
		if err != nil {
			return nil, err
		}
		showable = port.Services(svcs)
	case port.ShowTypeMessage:
		msgs, err := env.Messages()
		if err != nil {
			return nil, err
		}
		showable = port.Messages(msgs)
	case port.ShowTypeRPC:
		rpcs, err := env.RPCs()
		if err != nil {
			return nil, err
		}
		showable = port.RPCs(rpcs)
	case port.ShowTypeHeader:
		headers := env.Headers()
		showable = port.Headers(headers)
	default:
		return nil, errors.New("unknown showable type")
	}

	return outputPort.Show(showable)
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
		row := []string{f.Name(), f.Type()}
		rows = append(rows, row)
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}
