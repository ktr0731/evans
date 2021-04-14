package repl

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/usecase"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/types/descriptorpb"
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
	fields := td.(*desc.MessageDescriptor).GetFields()
	rows := make([][]string, len(fields))
	for i, field := range fields {
		rows[i] = []string{
			field.GetName(),
			presentTypeName(field),
			strconv.FormatBool(field.IsRepeated() && !field.IsMap()),
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	table.AppendBulk(rows)
	table.Render()
	return nil
}

func presentTypeName(f *desc.FieldDescriptor) string {
	typeName := f.GetType().String()

	switch f.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if f.IsMap() {
			typeName = fmt.Sprintf(
				"map<%s, %s>",
				presentTypeName(f.GetMapKeyType()),
				presentTypeName(f.GetMapValueType()))
		} else {
			typeName += fmt.Sprintf(" (%s)", f.GetMessageType().GetName())
		}
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		typeName += fmt.Sprintf(" (%s)", f.GetEnumType().GetName())
	}
	return typeName
}
