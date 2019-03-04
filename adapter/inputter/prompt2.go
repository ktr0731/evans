package inputter

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/prompt"
	"github.com/ktr0731/evans/color"
	"github.com/ktr0731/evans/entity/env"
	"github.com/pkg/errors"
)

type PromptInputter2 struct {
	prompt       prompt.Prompt
	prefixFormat string
	env          env.Environment

	ancestor                       []string
	hasAncestorAndHasRepeatedField bool
}

func NewPromptV2(prefixFormat string, env env.Environment) *PromptInputter2 {
	return &PromptInputter2{
		prompt:       prompt.New(nil, nil),
		prefixFormat: prefixFormat,
		env:          env,
	}
}

func (i *PromptInputter2) Input(req *desc.MessageDescriptor) (proto.Message, error) {
	return i.inputMessage(req)
}

// inputMessage walks fields of the passed message descriptor and sets inputted values.
//
func (i *PromptInputter2) inputMessage(d *desc.MessageDescriptor) (proto.Message, error) {
	dmsg := dynamic.NewMessage(d)

	oneofs := map[string][]*desc.FieldDescriptor{}
	// TODO: set color
	for _, oneof := range d.GetOneOfs() {
		oneofs[oneof.GetName()] = oneof.GetChoices()
	}

	// topLevelMsgs := map[string]proto.Message{}

	inputtedOneofs := map[string]interface{}{}
	for _, field := range d.GetFields() {
		inputField(i.prompt, dmsg, field, oneofs, inputtedOneofs)
	}
	return nil, nil
}

// fieldInputter tries to input values to a field.
type fieldInputter2 struct {
	f *desc.FieldDescriptor
}

func newFieldInputter2(
	prompt prompt.Prompt,
	prefixFormat string,
	ancestor []string,
	hasAncestorAndHasRepeatedField bool,
	hasDirectCycledParent bool,
	color color.Color,
) *fieldInputter2 {
	return &
}

func inputField(
	prompt prompt.Prompt,
	dmsg *dynamic.Message,
	field *desc.FieldDescriptor,
	oneofs map[string][]*desc.FieldDescriptor,
	inputtedOneofs map[string]interface{},
) error {
	if field.IsRepeated() {
		panic("TODO: repeated")
		for {
			// do
		}
	} else {
		// TODO: oneof
	}
	if isOneOfChoiceField(field) {
		oneOfName := field.GetOneOf().GetFullyQualifiedName()
		if _, found := inputtedOneofs[oneOfName]; found {
			fmt.Println("SKIP")
			return nil
		}
		choices := field.GetOneOf().GetChoices()
		choiceStrs := make([]string, len(choices))
		for i, choice := range choices {
			choiceStrs[i] = choice.GetName()
		}
		res, err := prompt.Select(field.GetOneOf().GetName(), choiceStrs)
		if err != nil {
			return errors.Wrap(err, "failed to select oneof field")
		}

		// Names of oneof fields must be unique, so it doesn't need to use FQN.
		for _, choice := range choices {
			if res == choice.GetName() {
				err := inputField(nil, field)
				if err != nil {
					return errors.Wrapf(err, "failed to input field '%s'", field.GetFullyQualifiedName())
				}
				inputtedOneofs[oneOfName] = nil
			}
		}
	} else {
		prompt.SetPrefix(i.makePrefix(field))
		in, err := prompt.Input()
		if err != nil {
			return nil, errors.Wrap(err, "failed to read user input")
		}
		fmt.Printf("%s (%s)\n", field.GetFullyQualifiedName(), field.GetType().String())
		dmsg.TrySetField(field, in)
	}
}

func isOneOfChoiceField(f *desc.FieldDescriptor) bool {
	return f.GetOneOf() != nil
}

// makePrefix makes prefix for field f.
func (i *PromptInputter2) makePrefix(f *desc.FieldDescriptor) string {
	return makePrefix2(i.prefixFormat, f, i.ancestor, i.hasAncestorAndHasRepeatedField)
}

const (
	repeatedStr       = "<repeated> "
	ancestorDelimiter = "::"
)

func makePrefix2(s string, f *desc.FieldDescriptor, ancestor []string, ancestorHasRepeated bool) string {
	joinedAncestor := strings.Join(ancestor, ancestorDelimiter)
	if joinedAncestor != "" {
		joinedAncestor += ancestorDelimiter
	}

	s = strings.Replace(s, "{ancestor}", joinedAncestor, -1)
	s = strings.Replace(s, "{name}", f.GetName(), -1)
	s = strings.Replace(s, "{type}", f.GetType().String(), -1)

	if f.IsRepeated() || ancestorHasRepeated {
		return repeatedStr + s
	}
	return s
}
