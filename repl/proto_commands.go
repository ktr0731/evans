package repl

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/ktr0731/evans/usecase"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type descCommand struct{}

func (c *descCommand) Synopsis() string {
	return "describe the structure of selected message"
}

func (c *descCommand) Help() string {
	return "usage: desc <message name>"
}

func (c *descCommand) FlagSet() (*pflag.FlagSet, bool) {
	return nil, false
}

func (c *descCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errArgumentRequired
	}
	return nil
}

func (c *descCommand) Run(w io.Writer, args []string) error {
	td, err := usecase.GetTypeDescriptor(args[0])
	if err != nil {
		return errors.Wrap(err, "failed to get the type descriptor")
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"field", "type", "repeated"})
	fields := td.(protoreflect.MessageDescriptor).Fields()
	rows := make([][]string, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		rows[i] = []string{
			string(field.Name()),
			presentTypeName(field),
			strconv.FormatBool(field.IsList() && !field.IsMap()), // TODO
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	table.AppendBulk(rows)
	table.Render()
	return nil
}

func presentTypeName(f protoreflect.FieldDescriptor) string {
	typeName := f.Kind().String()

	switch f.Kind() {
	case protoreflect.MessageKind:
		if f.IsMap() {
			typeName = fmt.Sprintf(
				"map<%s, %s>",
				presentTypeName(f.MapKey()),
				presentTypeName(f.MapValue()))
		} else {
			typeName += fmt.Sprintf(" (%s)", f.Message().Name())
		}
	case protoreflect.EnumKind:
		typeName += fmt.Sprintf(" (%s)", f.Enum().Name())
	}
	return typeName
}
