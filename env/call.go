package env

import (
	"fmt"
	"io"
	"strconv"

	"github.com/k0kubun/pp"
	"github.com/lycoris0731/evans/model"
	"github.com/peterh/liner"
)

// msg なら再帰的構造になる
type field struct {
	isPrimitive bool
	name        string
	pVal        *string
	mVal        *field
	fType       string
}

func (e *Env) Call(name string) (string, error) {
	rpc, err := e.GetRPC(name)
	if err != nil {
		return "", err
	}

	req, err := e.GetMessage(rpc.RequestType)
	if err != nil {
		return "", err
	}

	// res, err := e.GetMessage(rpc.ResponseType)
	// if err != nil {
	// 	return "", err
	// }

	inputs, err := inputFields(req.Fields)
	if err != nil {
		return "", err
	}

	pp.Println(inputs)

	return "", nil
}

func inputFields(fields []*model.Field) ([]*field, error) {
	const format = "%s (%s) -> "

	liner := liner.NewLiner()
	defer liner.Close()

	input := make([]*field, len(fields))
	max := maxLen(fields, format)
	for i, f := range fields {

		// TODO: msg
		// if descriptor.FieldDescriptorProto_Type_name[f.Type] == "TYPE_MESSAGE" {
		// 	// inputFields()
		// }
		l, err := liner.Prompt(fmt.Sprintf("%"+strconv.Itoa(max)+"s", fmt.Sprintf(format, f.Name, f.Type)))
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		input[i] = &field{
			name:        f.JSONName,
			isPrimitive: true,
			pVal:        &l,
			fType:       f.Type.String(),
		}
	}
	return input, nil
}

func maxLen(fields []*model.Field, format string) int {
	var max int
	for _, f := range fields {
		l := len(fmt.Sprintf(format, f.JSONName, f.Type))
		if l > max {
			max = l
		}
	}
	return max
}
