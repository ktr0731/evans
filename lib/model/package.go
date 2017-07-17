package model

import (
	"bytes"

	"github.com/olekukonko/tablewriter"
)

type Packages []string

func (p Packages) String() string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"package"})
	rows := [][]string{}
	for _, pack := range p {
		row := []string{pack}
		rows = append(rows, row)
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}
