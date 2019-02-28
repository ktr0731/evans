package inputter

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/adapter/prompt"
	"github.com/ktr0731/evans/entity/env"
)

type PromptInputter2 struct {
	prompt       prompt.Prompt
	prefixFormat string
	env          env.Environment
}

func NewPromptV2(prefixFormat string, env env.Environment) *PromptInputter2 {
	return &PromptInputter2{
		prompt:       prompt.New(nil, nil),
		prefixFormat: prefixFormat,
		env:          env,
	}
}

func (i *PromptInputter2) Input(req *desc.MessageDescriptor) (proto.Message, error) {
	inputtedOneOfs := map[string]interface{}{}
	// TODO: set color
	// for _, oneof := range req.GetOneOfs() {
	// 	fmt.Printf("--- oneof %s ---\n", oneof.GetName())
	// 	for _, choice := range oneof.GetChoices() {
	// 		fmt.Printf("* %s\n", choice.GetName())
	// 	}
	// }
	// fmt.Println()

	// topLevelMsgs := map[string]proto.Message{}

	for _, field := range req.GetFields() {
		// if field.IsRepeated() {
		// 	for {
		// 		// do
		// 	}
		// } else {
		// 	// TODO: oneof
		// }
		if isOneOfChoiceField(field) {
			oneOfName := field.GetOneOf().GetFullyQualifiedName()
			if _, found := inputtedOneOfs[oneOfName]; found {
				fmt.Println("SKIP")
				continue
			}
			fmt.Printf("--- oneof %s ---\n", field.GetName())
			for _, choice := range field.GetOneOf().GetChoices() {
				fmt.Printf("* %s\n", choice.GetName())
			}
			fmt.Printf("%s (%s) oneof = %s\n", field.GetFullyQualifiedName(), field.GetType().String(), field.GetOneOf().GetFullyQualifiedName())
			inputtedOneOfs[oneOfName] = nil
		} else {
			fmt.Printf("%s (%s)\n", field.GetFullyQualifiedName(), field.GetType().String())
		}
	}
	return nil, nil
}

func isOneOfChoiceField(f *desc.FieldDescriptor) bool {
	return f.GetOneOf() != nil
}
