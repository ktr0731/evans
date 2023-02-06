package proto

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/prompt"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/descriptorpb"
)

// InteractiveFiller is an implementation of fill.InteractiveFiller.
// It let you input request fields interactively.
type InteractiveFiller struct {
	prompt       prompt.Prompt
	prefixFormat string
}

// NewInteractiveFiller instantiates a new filler that fills each field interactively.
func NewInteractiveFiller(prompt prompt.Prompt, prefixFormat string) *InteractiveFiller {
	return &InteractiveFiller{
		prompt:       prompt,
		prefixFormat: prefixFormat,
	}
}

// Fill receives v that is an instance of *dynamic.Message.
// Fill let you input each field interactively by using a prompt. v will be set field values inputted by a prompt.
//
// Note that Fill resets the previous state when it is called again.
func (f *InteractiveFiller) Fill(v interface{}, opts fill.InteractiveFillerOpts) error {
	msg, ok := v.(*dynamic.Message)
	if !ok {
		return fill.ErrCodecMismatch
	}

	resolver := newResolver(f.prompt, f.prefixFormat, prompt.ColorInitial, msg, nil, false, opts)
	_, err := resolver.resolve()
	if err != nil {
		return err
	}

	return nil
}

type resolver struct {
	prompt       prompt.Prompt
	prefixFormat string
	color        prompt.Color

	msg *dynamic.Message

	m         *desc.MessageDescriptor
	ancestors []string
	// repeated represents that the message is repeated field or not.
	// If the message is not a field or not a repeated field, it is false.
	repeated bool

	opts fill.InteractiveFillerOpts
}

func newResolver(
	prompt prompt.Prompt,
	prefixFormat string,
	color prompt.Color,
	msg *dynamic.Message,
	ancestors []string,
	repeated bool,
	opts fill.InteractiveFillerOpts,
) *resolver {
	return &resolver{
		prompt:       prompt,
		prefixFormat: prefixFormat,
		color:        color,
		msg:          msg,
		m:            msg.GetMessageDescriptor(),
		ancestors:    ancestors,
		repeated:     repeated,
		opts:         opts,
	}
}

func (r *resolver) resolve() (*dynamic.Message, error) {
	selectedOneof := make(map[string]interface{})

	for _, f := range r.m.GetFields() {
		if isOneOfField := f.GetOneOf() != nil; isOneOfField {
			fqn := f.GetOneOf().GetFullyQualifiedName()
			if _, selected := selectedOneof[fqn]; selected {
				// Skip if one of choices is already selected.
				continue
			}

			selectedOneof[fqn] = nil
			if err := r.resolveOneof(f.GetOneOf()); err != nil {
				return nil, err
			}
			continue
		}

		err := r.resolveField(f)
		if errors.Is(err, prompt.ErrSkip) {
			continue
		}
		if errors.Is(err, prompt.ErrAbort) {
			return r.msg, nil
		}
		if err != nil {
			return nil, err
		}
	}

	return r.msg, nil
}

func (r *resolver) resolveOneof(o *desc.OneOfDescriptor) error {
	choices := make([]string, 0, len(o.GetChoices()))
	for _, c := range o.GetChoices() {
		choices = append(choices, c.GetName())
	}

	choice, err := r.selectChoices(o.GetFullyQualifiedName(), choices)
	if err != nil {
		return err
	}

	return r.resolveField(o.GetChoices()[choice])
}

func (r *resolver) resolveField(f *desc.FieldDescriptor) error {
	resolve := func(f *desc.FieldDescriptor) (interface{}, error) {
		var converter func(string) (interface{}, error)

		switch t := f.GetType(); t {
		case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
			if r.skipMessage(f) {
				return nil, prompt.ErrSkip
			}

			msgr := newResolver(
				r.prompt,
				r.prefixFormat,
				r.color.NextVal(),
				dynamic.NewMessage(f.GetMessageType()),
				append(r.ancestors, f.GetName()),
				r.repeated || f.IsRepeated(),
				r.opts,
			)
			return msgr.resolve()
		case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
			return r.resolveEnum(r.makePrefix(f), f.GetEnumType())
		case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
			converter = func(v string) (interface{}, error) { return strconv.ParseFloat(v, 64) }

		case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
			converter = func(v string) (interface{}, error) {
				f, err := strconv.ParseFloat(v, 32)
				return float32(f), err
			}

		case descriptorpb.FieldDescriptorProto_TYPE_INT64,
			descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
			descriptorpb.FieldDescriptorProto_TYPE_SINT64:
			converter = func(v string) (interface{}, error) { return strconv.ParseInt(v, 10, 64) }

		case descriptorpb.FieldDescriptorProto_TYPE_UINT64,
			descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
			converter = func(v string) (interface{}, error) { return strconv.ParseUint(v, 10, 64) }

		case descriptorpb.FieldDescriptorProto_TYPE_INT32,
			descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
			descriptorpb.FieldDescriptorProto_TYPE_SINT32:
			converter = func(v string) (interface{}, error) {
				i, err := strconv.ParseInt(v, 10, 32)
				return int32(i), err
			}

		case descriptorpb.FieldDescriptorProto_TYPE_UINT32,
			descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
			converter = func(v string) (interface{}, error) {
				u, err := strconv.ParseUint(v, 10, 32)
				return uint32(u), err
			}

		case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
			converter = func(v string) (interface{}, error) { return strconv.ParseBool(v) }

		case descriptorpb.FieldDescriptorProto_TYPE_STRING:
			converter = func(v string) (interface{}, error) { return v, nil }

		// Use strconv.Unquote to interpret byte literals and Unicode literals.
		// For example, a user inputs `\x6f\x67\x69\x73\x6f`,
		// His expects "ogiso" in string, but backslashes in the input are not interpreted as an escape sequence.
		// So, we need to call strconv.Unquote to interpret backslashes as an escape sequence.
		case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
			converter = func(v string) (interface{}, error) {
				if r.opts.BytesAsBase64 {
					b, err := base64.StdEncoding.DecodeString(v)
					if err == nil {
						return b, nil
					}
				} else if r.opts.BytesFromFile {
					b, err := os.ReadFile(v)
					if err == nil {
						return b, nil
					}
				}

				v, err := strconv.Unquote(`"` + v + `"`)
				return []byte(v), err
			}

		default:
			return nil, fmt.Errorf("invalid type: %s", t)
		}

		prefix := r.makePrefix(f)

		return r.input(prefix, f, converter)
	}

	if !f.IsRepeated() {
		v, err := resolve(f)
		if err != nil {
			return err
		}

		return r.msg.TrySetField(f, v)
	}

	color := r.color

	for {
		// Return nil to keep inputted values.
		if !r.addRepeatedField(f) {
			return nil
		}

		r.prompt.SetPrefixColor(color)
		color.Next()

		v, err := resolve(f)
		if err == io.EOF {
			// io.EOF signals the end of inputting repeated field.
			// Return nil to keep inputted values.
			return nil
		}
		if err != nil {
			return err
		}

		if err := r.msg.TryAddRepeatedField(f, v); err != nil {
			return err
		}
	}
}

func (r *resolver) resolveEnum(prefix string, e *desc.EnumDescriptor) (int32, error) {
	choices := make([]string, 0, len(e.GetValues()))
	for _, v := range e.GetValues() {
		choices = append(choices, v.GetName())
	}

	choice, err := r.selectChoices(prefix, choices)
	if err != nil {
		return 0, err
	}

	value := e.GetValues()[choice].AsEnumValueDescriptorProto()

	return *value.Number, nil
}

func (r *resolver) input(prefix string, f *desc.FieldDescriptor, converter func(string) (interface{}, error)) (interface{}, error) {
	r.prompt.SetPrefix(prefix)
	r.prompt.SetPrefixColor(r.color)

	in, err := r.prompt.Input()
	if err != nil {
		return nil, err
	}
	if in == "" {
		if f.IsRepeated() {
			builder, err := builder.FromField(f)
			if err != nil {
				return nil, err
			}

			// Clear "repeated".
			builder.Label = descriptorpb.FieldDescriptorProto_Label(0)
			f, err = builder.Build()
			if err != nil {
				return nil, err
			}
		}
		return f.GetDefaultValue(), nil
	}

	return converter(in)
}

func (r *resolver) selectChoices(msg string, choices []string) (int, error) {
	n, _, err := r.prompt.Select(msg, choices)
	if errors.Is(err, prompt.ErrAbort) {
		// Skip inputting and use default value.
		return 0, nil
	}
	if errors.Is(err, io.EOF) {
		return 0, io.EOF
	}
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (r *resolver) addRepeatedField(f *desc.FieldDescriptor) bool {
	if !r.opts.AddRepeatedManually {
		if f.GetType() != descriptorpb.FieldDescriptorProto_TYPE_MESSAGE || len(f.GetMessageType().GetFields()) != 0 {
			return true
		}

		// f is repeated empty message field. It will cause infinite-loop if r.opts.AddRepeatedManually is false.
		// For user's experience, always display prompt in this case.
	}

	msg := fmt.Sprintf("add a repeated field value? field=%s", f.GetFullyQualifiedName())
	choices := []string{"yes", "no"}
	n, _, err := r.prompt.Select(msg, choices)
	if err != nil || n == 1 {
		return false
	}

	return true
}

func (r *resolver) skipMessage(f *desc.FieldDescriptor) bool {
	if !r.opts.DigManually {
		return false
	}

	msg := fmt.Sprintf("dig down? field=%s", f.GetFullyQualifiedName())
	n, _, _ := r.prompt.Select(msg, []string{"dig down", "skip"})
	return n == 1
}

func (r *resolver) makePrefix(field *desc.FieldDescriptor) string {
	const delimiter = "::"

	joinedAncestor := strings.Join(r.ancestors, delimiter)
	if joinedAncestor != "" {
		joinedAncestor += delimiter
	}

	s := r.prefixFormat

	s = strings.ReplaceAll(s, "{ancestor}", joinedAncestor)
	s = strings.ReplaceAll(s, "{name}", field.GetName())
	s = strings.ReplaceAll(s, "{type}", field.GetType().String())

	if r.repeated || field.IsRepeated() {
		return "<repeated> " + s
	}

	return s
}
