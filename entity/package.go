package entity

import (
	"bytes"

	"github.com/olekukonko/tablewriter"
)

type Package struct {
	Name     string
	Services Services
	Messages Messages
}

// GetMessage is only used to get a root message
func (p *Package) GetMessage(name string) (*Message, error) {
	// nested message は辿る必要がない
	for _, msg := range p.Messages {
		if msg.Name == name {
			return msg, nil
		}
	}
	return nil, ErrInvalidMessageName
}

type Packages []*Package

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
