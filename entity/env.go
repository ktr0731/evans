package entity

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/AlecAivazis/survey"
	prompt "github.com/c-bata/go-prompt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/parser"
	"github.com/pkg/errors"
)

var (
	ErrPackageUnselected    = errors.New("package unselected")
	ErrServiceUnselected    = errors.New("service unselected")
	ErrUnknownTarget        = errors.New("unknown target")
	ErrUnknownPackage       = errors.New("unknown package")
	ErrUnknownService       = errors.New("unknown service")
	ErrInvalidServiceName   = errors.New("invalid service name")
	ErrInvalidMessageName   = errors.New("invalid message name")
	ErrInvalidRPCName       = errors.New("invalid RPC name")
	ErrServiceCachingFailed = errors.New("service caching failed")
)

type Environment interface {
	Packages() Packages
	Services() (Services, error)
	Messages() (Messages, error)
	RPCs() (RPCs, error)
	Service(name string) (*Service, error)
	Message(name string) (*Message, error)
	RPC(name string) (*RPC, error)

	UsePackage(name string) error
	UseService(name string) error
}

// pkgList is used by showing all packages
// pkg is used by extract a package by package name
type cache struct {
	pkgList Packages
	pkg     map[string]*Package
}

type state struct {
	currentPackage string
	currentService string
}

type Env struct {
	desc  *parser.FileDescriptorSet
	state state

	config *config.Env

	cache cache
}

func New(desc *parser.FileDescriptorSet, config *config.Env) (*Env, error) {
	return &Env{
		desc:   desc,
		config: config,
		cache: cache{
			pkg: map[string]*Package{},
		},
	}, nil
}

func (e *Env) HasCurrentPackage() bool {
	return e.state.currentPackage != ""
}

func (e *Env) HasCurrentService() bool {
	return e.state.currentService != ""
}

func (e *Env) Packages() Packages {
	if e.cache.pkgList != nil {
		return e.cache.pkgList
	}

	pkgNames := e.desc.GetPackages()
	pkgs := make(Packages, len(pkgNames))
	for i, name := range pkgNames {
		pkgs[i] = &Package{Name: name}
	}

	e.cache.pkgList = pkgs

	return pkgs
}

func (e *Env) Services() (Services, error) {
	if !e.HasCurrentPackage() {
		return nil, ErrPackageUnselected
	}

	// services, messages and rpc are cached to e.cache when called UsePackage()
	// if messages isn't cached, it occurred panic
	return e.cache.pkg[e.state.currentPackage].Services, nil
}

func (e *Env) Messages() (Messages, error) {
	// TODO: current package 以外からも取得したい
	if !e.HasCurrentPackage() {
		return nil, ErrPackageUnselected
	}

	// same as GetServices()
	return e.cache.pkg[e.state.currentPackage].Messages, nil
}

func (e *Env) RPCs() (RPCs, error) {
	if !e.HasCurrentService() {
		return nil, ErrServiceUnselected
	}

	svc, err := e.Service(e.state.currentService)
	if err != nil {
		return nil, err
	}
	return svc.RPCs, nil
}

func (e *Env) Service(name string) (*Service, error) {
	svc, err := e.Services()
	if err != nil {
		return nil, err
	}
	for _, svc := range svc {
		if name == svc.Name {
			return svc, nil
		}
	}
	return nil, errors.Wrapf(ErrInvalidServiceName, "%s not found", name)
}

func (e *Env) Message(name string) (*Message, error) {
	// Person2 で panic
	msg, err := e.Messages()
	if err != nil {
		return nil, err
	}
	for _, msg := range msg {
		msgName := e.getNameFromFQN(name)
		if msgName == msg.Name {
			return msg, nil
		}
	}
	return nil, errors.Wrapf(ErrInvalidMessageName, "%s not found", name)
}

func (e *Env) RPC(name string) (*RPC, error) {
	rpcs, err := e.RPCs()
	if err != nil {
		return nil, err
	}
	for _, rpc := range rpcs {
		if name == rpc.Name {
			return rpc, nil
		}
	}
	return nil, errors.Wrapf(ErrInvalidRPCName, "%s not found", name)
}

func (e *Env) UsePackage(name string) error {
	for _, p := range e.desc.GetPackages() {
		if name == p {
			e.state.currentPackage = name
			return e.loadPackage(p)
		}
	}
	return errors.Wrapf(ErrUnknownPackage, "%s not found", name)
}

func (e *Env) UseService(name string) error {
	// set extracted package if passed service which has package name
	if e.state.currentPackage == "" {
		s := strings.SplitN(name, ".", 2)
		if len(s) != 2 {
			return errors.Wrap(ErrPackageUnselected, "please set package (package_name.service_name or set --package flag)")
		}
		if err := e.UsePackage(s[0]); err != nil {
			return errors.Wrapf(err, name)
		}
	}
	services, err := e.Services()
	if err != nil {
		return errors.Wrapf(err, "failed to get services")
	}
	for _, svc := range services {
		if name == svc.Name {
			e.state.currentService = name
			return nil
		}
	}
	return errors.Wrapf(ErrUnknownService, "%s not found", name)
}

func (e *Env) GetDSN() string {
	if e.state.currentPackage == "" {
		return ""
	}
	dsn := e.state.currentPackage
	if e.state.currentService != "" {
		dsn += "." + e.state.currentService
	}
	return dsn
}

// loadPackage loads all services and messages in itself
func (e *Env) loadPackage(name string) error {
	// prevent duplicated loading
	_, ok := e.cache.pkg[name]
	if ok {
		return nil
	}

	dSvc := e.desc.GetServices(name)
	dMsg := e.desc.GetMessages(name)

	// Messages: actual message size is greater than or equal to len(dMsg)
	//           because message can be contain other messages as a field
	e.cache.pkg[name] = &Package{
		Name:     name,
		Services: make(Services, len(dSvc)),
		Messages: make(Messages, 0, len(dMsg)),
	}

	services := make(Services, len(dSvc))
	for i, svc := range dSvc {
		services[i] = NewService(svc)
		services[i].RPCs = NewRPCs(svc)
	}
	e.cache.pkg[name].Services = services

	messages := make(Messages, len(dMsg))
	for i, msg := range dMsg {
		messages[i] = NewMessage(msg)

		fields, err := NewFields(e.cache.pkg[name], messages[i])
		if err != nil {
			return errors.Wrapf(err, "failed to get field of %s", msg.GetName())
		}
		messages[i].Fields = fields

		// cache each result by each time because some messages depends on some messages
		e.cache.pkg[name].Messages = append(e.cache.pkg[name].Messages, messages[i])
	}

	return nil
}

// Full Qualified Name
// It contains message or service with package name
// e.g.: .test.Person -> Person
func (e *Env) getNameFromFQN(fqn string) string {
	return strings.TrimLeft(fqn, "."+e.state.currentPackage+".")
}

// getMessage is a closure which has current states
// it is passed by entity.NewField() for get message from current package
func (e *Env) getMessage() func(typeName string) (*Message, error) {
	return func(msgName string) (*Message, error) {
		return e.Message(msgName)
	}
}

func (e *Env) getService() func(typeName string) (*Service, error) {
	return func(svcName string) (*Service, error) {
		return e.Service(svcName)
	}
}

/*
 * call
 */

// fieldable is only used to set primitive, enum, oneof fields.
type fieldable interface {
	fieldable()
	isNil() bool
}

type baseField struct {
	name     string
	descType descriptor.FieldDescriptorProto_Type
	desc     *desc.FieldDescriptor
}

func newBaseField(f *desc.FieldDescriptor) *baseField {
	return &baseField{
		name:     f.GetName(),
		desc:     f,
		descType: f.GetType(),
	}
}

func (f *baseField) fieldable() {}

// primitiveField is used to read and store input for each primitiveField
type primitiveField struct {
	*baseField
	val string
}

func (f *primitiveField) isNil() bool {
	return f.val == ""
}

type messageField struct {
	*baseField
	val []fieldable
}

func (f *messageField) isNil() bool {
	return f.val == nil
}

type repeatedField struct {
	*baseField
	val []fieldable
}

func (f *repeatedField) isNil() bool {
	return len(f.val) == 0
}

type enumField struct {
	*baseField
	val *desc.EnumValueDescriptor
}

func (f *enumField) isNil() bool {
	return f.val == nil
}

// Call calls a RPC which is selected
// RPC is called after inputting field values interactively
func (e *Env) Call(name string) (string, error) {
	rpc, err := e.RPC(name)
	if err != nil {
		return "", err
	}

	// TODO: GetFields は OneOf の要素まで取得してしまう
	input, err := e.inputFields([]string{}, rpc.RequestType, prompt.DarkGreen)
	if errors.Cause(err) == io.EOF {
		return "", nil
	} else if err != nil {
		return "", err
	}

	req := dynamic.NewMessage(rpc.RequestType)
	if err = e.setInput(req, input); err != nil {
		return "", err
	}

	res := dynamic.NewMessage(rpc.ResponseType)
	conn, err := connect(e.config.Server)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	ep := e.genEndpoint(name)
	md := metadata.MD{}
	for _, h := range e.config.Request.Header {
		md[h.Key] = []string{h.Value}
	}
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	if err := grpc.Invoke(ctx, ep, req, res, conn); err != nil {
		return "", err
	}

	out, err := formatOutput(res)
	if err != nil {
		return "", err
	}

	return out, nil
}

func (e *Env) CallWithScript(input io.Reader, rpcName string) error {
	rpc, err := e.RPC(rpcName)
	if err != nil {
		return err
	}
	req := dynamic.NewMessage(rpc.RequestType)
	if err := jsonpb.Unmarshal(input, req); err != nil {
		return err
	}
	res := dynamic.NewMessage(rpc.ResponseType)

	if err := e.call(e.genEndpoint(rpcName), req, res); err != nil {
		return err
	}

	out, err := formatOutput(res)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, out)

	return nil
}

// req, res は既に値が入っている前提
func (e *Env) call(endpoint string, req, res proto.Message) error {
	conn, err := connect(e.config.Server)
	if err != nil {
		return err
	}
	defer conn.Close()

	return grpc.Invoke(context.Background(), endpoint, req, res, conn)
}

func (e *Env) genEndpoint(rpcName string) string {
	ep := fmt.Sprintf("/%s.%s/%s", e.state.currentPackage, e.state.currentService, rpcName)
	return ep
}

func (e *Env) setInput(req *dynamic.Message, fields []fieldable) error {
	for _, field := range fields {
		switch f := field.(type) {
		case *primitiveField:
			pv := f.val

			v, err := castPrimitiveType(f, pv)
			if err != nil {
				return err
			}
			if err := req.TrySetField(f.desc, v); err != nil {
				return err
			}

		case *messageField:
			// TODO
			msg := dynamic.NewMessage(f.desc.GetMessageType())
			if err := e.setInput(msg, f.val); err != nil {
				return err
			}
			req.SetField(f.desc, msg)
		case *repeatedField:
			// ここの f.desc に Add する

			if f.desc.GetMessageType() != nil {
				msg := dynamic.NewMessage(f.desc.GetMessageType())
				if err := e.setInput(msg, f.val); err != nil {
					return err
				}
				req.TryAddRepeatedField(f.desc, msg)
			} else { // primitive type
				for _, field := range f.val {
					f2 := field.(*primitiveField)
					v, err := castPrimitiveType(f2, f2.val)
					if err != nil {
						return err
					}
					if err := req.TryAddRepeatedField(f.desc, v); err != nil {
						return err
					}
				}
			}
		case *enumField:
			// TODO
			req.SetField(f.desc, f.val.GetNumber())
		}
	}
	return nil
}

func (e *Env) inputFields(ancestor []string, msg *desc.MessageDescriptor, color prompt.Color) ([]fieldable, error) {
	fields := msg.GetFields()

	input := make([]fieldable, 0, len(fields))
	max := maxLen(fields, e.config.InputPromptFormat)
	// TODO: ずれてる
	promptFormat := fmt.Sprintf("%"+strconv.Itoa(max)+"s", e.config.InputPromptFormat)

	inputField := e.fieldInputer(ancestor, promptFormat, color)

	encountered := map[string]map[string]bool{
		"oneof": map[string]bool{},
		"enum":  map[string]bool{},
	}
	for _, f := range fields {
		var in fieldable
		// message field, enum field or primitive field
		switch {
		case isOneOf(f):
			oneOf := f.GetOneOf()

			if encountered["oneof"][oneOf.GetFullyQualifiedName()] {
				continue
			}

			encountered["oneof"][oneOf.GetFullyQualifiedName()] = true

			opts := make([]string, len(oneOf.GetChoices()))
			optMap := map[string]*desc.FieldDescriptor{}
			for i, c := range oneOf.GetChoices() {
				opts[i] = c.GetName()
				optMap[c.GetName()] = c
			}

			var choice string
			err := survey.AskOne(&survey.Select{
				Message: oneOf.GetName(),
				Options: opts,
			}, &choice, nil)
			if err != nil {
				return nil, err
			}

			f = optMap[choice]
		case isEnumType(f):
			enum := f.GetEnumType()
			if encountered["enum"][enum.GetFullyQualifiedName()] {
				continue
			}

			encountered["enum"][enum.GetFullyQualifiedName()] = true

			opts := make([]string, len(enum.GetValues()))
			optMap := map[string]*desc.EnumValueDescriptor{}
			for i, o := range enum.GetValues() {
				opts[i] = o.GetName()
				optMap[o.GetName()] = o
			}

			var choice string
			err := survey.AskOne(&survey.Select{
				Message: enum.GetName(),
				Options: opts,
			}, &choice, nil)
			if err != nil {
				return nil, err
			}
			in = &enumField{
				baseField: newBaseField(f),
				val:       optMap[choice],
			}
		}

		if f.IsRepeated() {
			var repeated []fieldable
			// TODO: repeated であることを prompt に出したい
			for {
				s, err := inputField(f)
				if err != nil {
					return nil, err
				}
				if s.isNil() {
					break
				}
				repeated = append(repeated, s)
			}
			in = &repeatedField{
				baseField: newBaseField(f),
				val:       repeated,
			}
		} else if !isEnumType(f) {
			var err error
			in, err = inputField(f)
			if err != nil {
				return nil, err
			}
		}

		input = append(input, in)
	}
	return input, nil
}

func connect(config *config.Server) (*grpc.ClientConn, error) {
	// TODO: connection を使いまわしたい
	return grpc.Dial(fmt.Sprintf("%s:%s", config.Host, config.Port), grpc.WithInsecure())
}

func formatOutput(input proto.Message) (string, error) {
	m := jsonpb.Marshaler{Indent: "  "}
	out, err := m.MarshalToString(input)
	if err != nil {
		return "", err
	}
	return out + "\n", nil
}

// fieldInputer let us enter primitive or message field.
func (e *Env) fieldInputer(ancestor []string, promptFormat string, color prompt.Color) func(*desc.FieldDescriptor) (fieldable, error) {
	return func(f *desc.FieldDescriptor) (fieldable, error) {
		if isMessageType(f.GetType()) {
			fields, err := e.inputFields(append(ancestor, f.GetName()), f.GetMessageType(), color)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read inputs")
			}
			color = prompt.DarkGreen + (color+1)%16
			return &messageField{
				baseField: newBaseField(f),
				val:       fields,
			}, nil
		} else { // primitive
			promptStr := promptFormat
			ancestor := strings.Join(ancestor, e.config.AncestorDelimiter)
			if ancestor != "" {
				ancestor = "@" + ancestor
			}
			// TODO: text template
			promptStr = strings.Replace(promptStr, "{ancestor}", ancestor, -1)
			promptStr = strings.Replace(promptStr, "{name}", f.GetName(), -1)
			promptStr = strings.Replace(promptStr, "{type}", f.GetType().String(), -1)

			in := prompt.Input(
				promptStr,
				inputCompleter,
				prompt.OptionPrefixTextColor(color),
			)

			return &primitiveField{
				baseField: newBaseField(f),
				val:       in,
			}, nil
		}
	}
}

func castPrimitiveType(f *primitiveField, pv string) (interface{}, error) {
	// it holds value and error of conversion
	// each cast (Parse*) returns falsy value when failed to parse argument
	var v interface{}
	var err error

	switch f.descType {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		v, err = strconv.ParseFloat(pv, 64)

	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		v, err = strconv.ParseFloat(pv, 32)
		v = float32(v.(float64))

	case descriptor.FieldDescriptorProto_TYPE_INT64:
		v, err = strconv.ParseInt(pv, 10, 64)

	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		v, err = strconv.ParseUint(pv, 10, 64)

	case descriptor.FieldDescriptorProto_TYPE_INT32:
		v, err = strconv.ParseInt(f.val, 10, 32)
		v = int32(v.(int64))

	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		v, err = strconv.ParseUint(pv, 10, 32)
		v = uint32(v.(uint64))

	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		v, err = strconv.ParseUint(pv, 10, 64)

	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		v, err = strconv.ParseUint(pv, 10, 32)
		v = uint32(v.(uint64))

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		v, err = strconv.ParseBool(pv)

	case descriptor.FieldDescriptorProto_TYPE_STRING:
		// already string
		v = pv

	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		v = []byte(pv)

	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		v, err = strconv.ParseUint(pv, 10, 64)

	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		v, err = strconv.ParseUint(pv, 10, 32)
		v = int32(v.(int64))

	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		v, err = strconv.ParseInt(pv, 10, 64)

	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		v, err = strconv.ParseInt(pv, 10, 32)
		v = int32(v.(int64))

	default:
		return nil, fmt.Errorf("invalid type: %#v", f.descType)
	}
	return v, err
}

func maxLen(fields []*desc.FieldDescriptor, format string) int {
	var max int
	for _, f := range fields {
		if isMessageType(f.GetType()) {
			continue
		}
		prompt := format
		elems := map[string]string{
			"name": f.GetName(),
			"type": f.GetType().String(),
		}
		for k, v := range elems {
			prompt = strings.Replace(prompt, "{"+k+"}", v, -1)
		}
		l := len(format)
		if l > max {
			max = l
		}
	}
	return max
}

func isMessageType(typeName descriptor.FieldDescriptorProto_Type) bool {
	return typeName == descriptor.FieldDescriptorProto_TYPE_MESSAGE
}

func isOneOf(f *desc.FieldDescriptor) bool {
	return f.GetOneOf() != nil
}

func isEnumType(f *desc.FieldDescriptor) bool {
	return f.GetEnumType() != nil
}

func inputCompleter(d prompt.Document) []prompt.Suggest {
	return nil
}
