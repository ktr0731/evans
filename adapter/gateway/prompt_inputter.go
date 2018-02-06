package gateway

import (
	"strings"

	"github.com/AlecAivazis/survey"
	prompt "github.com/c-bata/go-prompt"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/pkg/errors"
)

var (
	ErrUnknownOneofFieldName = errors.New("unknown oneof field name")
	ErrUnknownEnumName       = errors.New("unknown enum name")
	EORF                     = errors.New("end of repeated field")
)

// for mocking
type prompter interface {
	Input() string
	Select(msg string, opts []string) (string, error)
	SetPrefix(prefix string) error
	SetPrefixColor(color prompt.Color) error
}

type RealPrompter struct {
	fieldPrompter *prompt.Prompt
}

func (p *RealPrompter) Input() string {
	return p.fieldPrompter.Input()
}

func (p *RealPrompter) Select(msg string, opts []string) (string, error) {
	var choice string
	err := survey.AskOne(&survey.Select{
		Message: msg,
		Options: opts,
	}, &choice, nil)
	return choice, err
}

func (p *RealPrompter) SetPrefix(prefix string) error {
	return prompt.OptionPrefix(prefix)(p.fieldPrompter)
}

func (p *RealPrompter) SetPrefixColor(color prompt.Color) error {
	return prompt.OptionPrefixTextColor(color)(p.fieldPrompter)
}

// mixin go-prompt
func NewPrompt(config *config.Config, env entity.Environment) *Prompt {
	executor := func(in string) {
		return
	}
	completer := func(d prompt.Document) []prompt.Suggest {
		return nil
	}
	return newPrompt(&RealPrompter{prompt.New(executor, completer)}, config, env)
}

type Prompt struct {
	prompt prompter
	config *config.Config
	env    entity.Environment
}

func newPrompt(prompt prompter, config *config.Config, env entity.Environment) *Prompt {
	return &Prompt{
		prompt: prompt,
		config: config,
		env:    env,
	}
}

// impl of port.Inputter
func (i *Prompt) Input(reqType entity.Message) (proto.Message, error) {
	setter := protobuf.NewMessageSetter(reqType)
	// req := dynamic.NewMessage(reqType)
	fields := reqType.Fields()

	// DarkGreen is the initial color
	return newFieldInputter(i.prompt, i.config.Env.InputPromptFormat, setter, true, prompt.DarkGreen).Input(fields)
}

// fieldInputter inputs each fields of req in interactively
// first fieldInputter is instantiated per one request
type fieldInputter struct {
	prompt prompter
	setter *protobuf.MessageSetter

	prefixFormat string

	encountered map[string]map[string]bool
	msg         *dynamic.Message
	fields      []*desc.FieldDescriptor
	dep         msgDep

	color prompt.Color

	isTopLevelMessage bool

	// enteredEmptyInput is used to terminate repeated field inputting
	// if input is empty and enteredEmptyInput is true, exit repeated input prompt
	enteredEmptyInput bool
}

type msgDep map[string]*desc.MessageDescriptor

// resolve dependencies of reqType
// func resolveMessageDependency(msg *desc.MessageDescriptor, dep msgDep, encountered map[string]bool) {
// 	if encountered[msg.GetFullyQualifiedName()] {
// 		return
// 	}
//
// 	dep[msg.GetFullyQualifiedName()] = msg
// 	for _, f := range msg.GetFields() {
// 		if entity.IsMessageType(f.GetType()) {
// 			resolveMessageDependency(f.GetMessageType(), dep, encountered)
// 		}
// 	}
// }

func newFieldInputter(prompter prompter, prefixFormat string, setter *protobuf.MessageSetter, isTopLevelMessage bool, color prompt.Color) *fieldInputter {
	// dep := msgDep{}
	// resolveMessageDependency(msgType, dep, map[string]bool{})
	return &fieldInputter{
		prompt:       prompter,
		setter:       setter,
		prefixFormat: prefixFormat,
		encountered: map[string]map[string]bool{
			"oneof": map[string]bool{},
			"enum":  map[string]bool{},
		},
		isTopLevelMessage: isTopLevelMessage,
		color:             color,
	}
}

// Input will call itself for nested messages
//
// e.g.
// message Foo {
// 	Bar bar = 1;
//	string baz = 2;
// }
//
// Input is called two times
// one for Foo's primitive fields
// another one for bar's primitive fields
func (i *fieldInputter) Input(fields []entity.Field) (proto.Message, error) {
	if err := i.prompt.SetPrefixColor(i.color); err != nil {
		return nil, err
	}

	for _, field := range fields {
		if field.IsRepeated() {
			for {
				if err := i.inputField(field); err == EORF {
					break
				} else if err != nil {
					return nil, err
				}
			}
			i.enteredEmptyInput = false
		} else {
			if err := i.inputField(field); err != nil {
				return nil, err
			}
		}
	}

	return i.setter.Done(), nil
}

func (i *fieldInputter) encounteredOneof(oneof entity.OneOfField) bool {
	encountered := i.encountered["oneof"][oneof.FQRN()]
	i.encountered["oneof"][oneof.FQRN()] = true
	return encountered
}

func (i *fieldInputter) chooseOneof(oneof entity.OneOfField) (entity.Field, error) {
	options := make([]string, len(oneof.Choices()))
	fieldOf := map[string]entity.Field{}
	for _, choice := range oneof.Choices() {
		options = append(options, choice.FieldName())
		fieldOf[choice.FieldName()] = choice
	}

	choice, err := i.prompt.Select(oneof.FieldName(), options)
	if err != nil {
		return nil, err
	}

	f, ok := fieldOf[choice]
	if !ok {
		return nil, errors.Wrap(ErrUnknownOneofFieldName, choice)
	}

	return f, nil
}

func (i *fieldInputter) encounteredEnum(enum entity.EnumField) bool {
	encountered := i.encountered["enum"][enum.FQRN()]
	i.encountered["enum"][enum.FQRN()] = true
	return encountered
}

func (i *fieldInputter) chooseEnum(enum entity.EnumField) (entity.EnumValue, error) {
	options := make([]string, 0, len(enum.Values()))
	valOf := map[string]entity.EnumValue{}
	for _, v := range enum.Values() {
		options = append(options, v.Name())
		valOf[v.Name()] = v
	}

	choice, err := i.prompt.Select(enum.Name(), options)
	if err != nil {
		return nil, err
	}

	c, ok := valOf[choice]
	if !ok {
		return nil, errors.Wrap(ErrUnknownEnumName, choice)
	}

	return c, nil
}

func (i *fieldInputter) inputField(field entity.Field) error {
	// if oneof, choose one from selection
	if oneof, ok := field.(entity.OneOfField); ok {
		if i.encounteredOneof(oneof) {
			return nil
		}
		var err error
		field, err = i.chooseOneof(oneof)
		if err != nil {
			return err
		}
	}

	switch f := field.(type) {
	case entity.EnumField:
		if i.encounteredEnum(f) {
			return nil
		}
		v, err := i.chooseEnum(f)
		if err != nil {
			return err
		}
		if err := i.setter.SetField(f, v.Number()); err != nil {
			return err
		}
	case entity.MessageField:
		// nestedFields := i.dep[field.GetMessageType().GetFullyQualifiedName()].GetFields()
		setter := protobuf.NewMessageSetter(f)
		fields := f.Fields()
		msg, err := newFieldInputter(i.prompt, i.prefixFormat, setter, false, i.color).Input(fields)
		if err != nil {
			return err
		}
		if err := i.setter.SetField(f, msg); err != nil {
			return err
		}
		// increment prompt color to next one
		i.color = (i.color + 1) % 16
	case entity.PrimitiveField:
		if err := i.prompt.SetPrefix(makePrefix(i.prefixFormat, field, i.isTopLevelMessage)); err != nil {
			return err
		}
		if err := i.inputPrimitiveField(f); err != nil {
			return err
		}
	default:
		panic("unknown type: " + field.PBType())
	}
	return nil
}

func (i *fieldInputter) inputPrimitiveField(f entity.PrimitiveField) error {
	in := i.prompt.Input()

	if in == "" {
		if i.enteredEmptyInput {
			return EORF
		}
		i.enteredEmptyInput = true
		// ignore the input
		return i.inputPrimitiveField(f)
	}

	v, err := protobuf.ConvertValue(in, f)
	if err != nil {
		return err
	}
	return i.setter.SetField(f, v)
}

func (i *fieldInputter) setField(req *dynamic.Message, field *desc.FieldDescriptor, v interface{}) error {
	if field.IsRepeated() {
		return req.TryAddRepeatedField(field, v)
	}

	return req.TrySetField(field, v)
}

// makePrefix makes prefix for field f.
// isTopLevelMessage is used to show passed f is a message and it is top-level message.
// for example, person field is a message, Person. and a part of BorrowBookRequest.
// also BorrowBookRequest is a top-level message.
//
// message BorrowBookRequest {
//  Person person = 1;
//   Book book = 2;
// }
//
func makePrefix(s string, f entity.PrimitiveField, isTopLevelMessage bool) string {
	// ancestor := []string{}
	// var d desc.Descriptor = f.GetParent()
	//
	// // if f is a top-level, message, exclude name of top-level message.
	// if isTopLevelMessage {
	// 	d = d.GetParent()
	// }
	//
	// for d != nil {
	// 	ancestor = append([]string{d.GetName()}, ancestor...)
	// 	d = d.GetParent()
	// }
	// // remove file name
	// ancestor = ancestor[1:]
	//
	// joinedAncestor := strings.Join(ancestor, "::")
	// if joinedAncestor != "" {
	// 	joinedAncestor += "::"
	// }
	// s = strings.Replace(s, "{ancestor}", joinedAncestor, -1)
	s = strings.Replace(s, "{name}", f.FieldName(), -1)
	s = strings.Replace(s, "{type}", f.PBType(), -1)
	return s
}
