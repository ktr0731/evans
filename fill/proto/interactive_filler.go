package proto

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/prompt"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
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
func (f *InteractiveFiller) Fill(v *dynamicpb.Message, opts fill.InteractiveFillerOpts) error {
	resolver := newResolver(f.prompt, f.prefixFormat, prompt.ColorInitial, v, nil, false, opts)
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

	msg *dynamicpb.Message

	m         protoreflect.MessageDescriptor
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
	msg *dynamicpb.Message,
	ancestors []string,
	repeated bool,
	opts fill.InteractiveFillerOpts,
) *resolver {
	return &resolver{
		prompt:       prompt,
		prefixFormat: prefixFormat,
		color:        color,
		msg:          msg,
		m:            msg.Descriptor(),
		ancestors:    ancestors,
		repeated:     repeated,
		opts:         opts,
	}
}

func (r *resolver) resolve() (*dynamicpb.Message, error) {
	selectedOneof := make(map[string]interface{})

	// for _, f := range r.m.Fields(). {
	for i := 0; i < r.m.Fields().Len(); i++ {
		f := r.m.Fields().Get(i)

		if isOneOfField := f.ContainingOneof() != nil; isOneOfField {
			fqn := string(f.ContainingOneof().FullName())
			if _, selected := selectedOneof[fqn]; selected {
				// Skip if one of choices is already selected.
				continue
			}

			selectedOneof[fqn] = nil
			if err := r.resolveOneof(f.ContainingOneof()); err != nil {
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

func (r *resolver) resolveOneof(o protoreflect.OneofDescriptor) error {
	choices := make([]string, 0, o.Fields().Len())
	for i := 0; i < o.Fields().Len(); i++ {
		c := o.Fields().Get(i)
		choices = append(choices, string(c.Name()))
	}

	choice, err := r.selectChoices(string(o.FullName()), choices)
	if err != nil {
		return err
	}

	return r.resolveField(o.Fields().Get(choice))
}

func (r *resolver) resolveField(f protoreflect.FieldDescriptor) error {
	resolve := func(f protoreflect.FieldDescriptor) (protoreflect.Value, error) {
		var converter func(string) (protoreflect.Value, error)

		switch t := f.Kind(); t {
		case protoreflect.MessageKind:
			if r.skipMessage(f) {
				return protoreflect.Value{}, prompt.ErrSkip
			}

			msgr := newResolver(
				r.prompt,
				r.prefixFormat,
				r.color.NextVal(),
				dynamicpb.NewMessage(f.Message()),
				append(r.ancestors, string(f.Name())),
				r.repeated || f.IsList(),
				r.opts,
			)
			msg, err := msgr.resolve()
			if err != nil {
				return protoreflect.Value{}, err
			}

			return protoreflect.ValueOf(msg), nil
		case protoreflect.EnumKind:
			v, err := r.resolveEnum(r.makePrefix(f), f.Enum())
			if err != nil {
				return protoreflect.Value{}, err
			}

			return protoreflect.ValueOf(protoreflect.EnumNumber(v)), nil
		case protoreflect.DoubleKind:
			converter = func(v string) (protoreflect.Value, error) {
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return protoreflect.Value{}, err
				}

				return protoreflect.ValueOf(f), nil
			}

		case protoreflect.FloatKind:
			converter = func(v string) (protoreflect.Value, error) {
				f, err := strconv.ParseFloat(v, 32)
				if err != nil {
					return protoreflect.Value{}, err
				}

				return protoreflect.ValueOf(float32(f)), nil
			}

		case protoreflect.Int64Kind, protoreflect.Sfixed64Kind, protoreflect.Sint64Kind:
			converter = func(v string) (protoreflect.Value, error) {
				n, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return protoreflect.Value{}, err
				}

				return protoreflect.ValueOf(n), nil
			}

		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			converter = func(v string) (protoreflect.Value, error) {
				n, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					return protoreflect.Value{}, err
				}

				return protoreflect.ValueOf(n), nil
			}

		case protoreflect.Int32Kind, protoreflect.Sfixed32Kind, protoreflect.Sint32Kind:
			converter = func(v string) (protoreflect.Value, error) {
				i, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					return protoreflect.Value{}, err
				}

				return protoreflect.ValueOf(int32(i)), err
			}

		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			converter = func(v string) (protoreflect.Value, error) {
				u, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					return protoreflect.Value{}, err
				}

				return protoreflect.ValueOf(uint32(u)), err
			}

		case protoreflect.BoolKind:
			converter = func(v string) (protoreflect.Value, error) {
				b, err := strconv.ParseBool(v)
				if err != nil {
					return protoreflect.Value{}, err
				}

				return protoreflect.ValueOf(b), nil
			}

		case protoreflect.StringKind:
			converter = func(v string) (protoreflect.Value, error) { return protoreflect.ValueOf(v), nil }

		// Use strconv.Unquote to interpret byte literals and Unicode literals.
		// For example, a user inputs `\x6f\x67\x69\x73\x6f`,
		// His expects "ogiso" in string, but backslashes in the input are not interpreted as an escape sequence.
		// So, we need to call strconv.Unquote to interpret backslashes as an escape sequence.
		case protoreflect.BytesKind:
			converter = func(v string) (protoreflect.Value, error) {
				if r.opts.BytesAsBase64 {
					b, err := base64.StdEncoding.DecodeString(v)
					if err == nil {
						return protoreflect.ValueOf(b), nil
					}
				} else if r.opts.BytesFromFile {
					b, err := os.ReadFile(v)
					if err == nil {
						return protoreflect.ValueOf(b), nil
					}
				}

				v, err := strconv.Unquote(`"` + v + `"`)
				if err != nil {
					return protoreflect.Value{}, err
				}

				return protoreflect.ValueOf([]byte(v)), nil
			}

		default:
			return protoreflect.Value{}, fmt.Errorf("invalid type: %s", t)
		}

		prefix := r.makePrefix(f)

		return r.input(prefix, f, converter)
	}

	if f.Cardinality() != protoreflect.Repeated { // TODO: or cardinality
		v, err := resolve(f)
		if err != nil {
			return err
		}

		// TODO: is it okay?
		r.msg.Set(f, v)
		return nil
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

		switch {
		case f.IsList():
			r.msg.Mutable(f).List().Append(v)
		case f.IsMap():
			key := v.Message().Get(v.Message().Descriptor().Fields().Get(0)).MapKey()
			val := v.Message().Get(v.Message().Descriptor().Fields().Get(1))
			r.msg.Mutable(f).Map().Set(key, val)
		}
	}
}

func (r *resolver) resolveEnum(prefix string, e protoreflect.EnumDescriptor) (int32, error) {
	choices := make([]string, 0, e.Values().Len())
	// for _, v := range e.GetValues() {
	for i := 0; i < e.Values().Len(); i++ {
		v := e.Values().Get(i)
		choices = append(choices, string(v.Name()))
	}

	choice, err := r.selectChoices(prefix, choices)
	if err != nil {
		return 0, err
	}

	num := int32(e.Values().Get(choice).Number())

	return num, nil
}

func (r *resolver) input(prefix string, f protoreflect.FieldDescriptor, converter func(string) (protoreflect.Value, error)) (protoreflect.Value, error) {
	r.prompt.SetPrefix(prefix)
	r.prompt.SetPrefixColor(r.color)

	in, err := r.prompt.Input()
	if err != nil {
		return protoreflect.Value{}, err
	}
	if in == "" {
		if f.IsList() {
			return defaultValueFromKind(f.Kind()), nil
		}
		return protoreflect.ValueOf(f.Default().Interface()), nil
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

func (r *resolver) addRepeatedField(f protoreflect.FieldDescriptor) bool {
	if !r.opts.AddRepeatedManually {
		if f.Kind() != protoreflect.MessageKind || f.Message().Fields().Len() != 0 {
			return true
		}

		// f is repeated empty message field. It will cause infinite-loop if r.opts.AddRepeatedManually is false.
		// For user's experience, always display prompt in this case.
	}

	msg := fmt.Sprintf("add a repeated field value? field=%s", f.FullName())
	choices := []string{"yes", "no"}
	n, _, err := r.prompt.Select(msg, choices)
	if err != nil || n == 1 {
		return false
	}

	return true
}

func (r *resolver) skipMessage(f protoreflect.FieldDescriptor) bool {
	if !r.opts.DigManually {
		return false
	}

	msg := fmt.Sprintf("dig down? field=%s", f.FullName())
	n, _, _ := r.prompt.Select(msg, []string{"dig down", "skip"})
	return n == 1
}

func (r *resolver) makePrefix(field protoreflect.FieldDescriptor) string {
	const delimiter = "::"

	joinedAncestor := strings.Join(r.ancestors, delimiter)
	if joinedAncestor != "" {
		joinedAncestor += delimiter
	}

	s := r.prefixFormat

	s = strings.ReplaceAll(s, "{ancestor}", joinedAncestor)
	s = strings.ReplaceAll(s, "{name}", string(field.Name()))
	s = strings.ReplaceAll(s, "{type}", field.Kind().String())

	if r.repeated || field.IsList() {
		return "<repeated> " + s // TODO: OK? Or should check cardinality?
	}

	return s
}

var protoDefaults = map[protoreflect.Kind]interface{}{
	protoreflect.DoubleKind:   float64(0),
	protoreflect.FloatKind:    float32(0),
	protoreflect.Int64Kind:    int64(0),
	protoreflect.Uint64Kind:   uint64(0),
	protoreflect.Int32Kind:    int32(0),
	protoreflect.Uint32Kind:   uint32(0),
	protoreflect.Fixed64Kind:  uint64(0),
	protoreflect.Fixed32Kind:  uint32(0),
	protoreflect.BoolKind:     false,
	protoreflect.StringKind:   "",
	protoreflect.BytesKind:    []byte{},
	protoreflect.Sfixed64Kind: int64(0),
	protoreflect.Sfixed32Kind: int32(0),
	protoreflect.Sint64Kind:   int64(0),
	protoreflect.Sint32Kind:   int32(0),
}

// convertValue converts a string input pv to protoreflect.Value.
func defaultValueFromKind(kind protoreflect.Kind) protoreflect.Value {
	return protoreflect.ValueOf(protoDefaults[kind])
}
