package proto

import (
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

type anyResolver struct {
	protoregistry.ExtensionTypeResolver
	descSource DescriptorSource
}

func NewAnyResolver(descSource DescriptorSource) interface {
	protoregistry.ExtensionTypeResolver
	protoregistry.MessageTypeResolver
} {
	return &anyResolver{
		descSource: descSource,
	}
}

func (r *anyResolver) FindMessageByName(m protoreflect.FullName) (protoreflect.MessageType, error) {
	d, err := r.descSource.FindSymbol(string(m))
	if err != nil {
		return nil, err
	}
	if errors.Is(err, errSymbolNotFound) {
		// Fallback to protoregistry.GlobalTypes.
		return protoregistry.GlobalTypes.FindMessageByName(m)
	}

	md := d.(protoreflect.MessageDescriptor) // TODO: handle "ok".

	return dynamicpb.NewMessageType(md), nil
}

func (r *anyResolver) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	n := strings.LastIndex(url, "/")
	if n != -1 {
		url = url[n+1:]
	}

	d, err := r.descSource.FindSymbol(url)
	if err != nil && !errors.Is(err, errSymbolNotFound) {
		return nil, err
	}
	if errors.Is(err, errSymbolNotFound) {
		// Fallback to protoregistry.GlobalTypes.
		return protoregistry.GlobalTypes.FindMessageByURL(url)
	}

	md := d.(protoreflect.MessageDescriptor) // TODO: handle "ok".

	return dynamicpb.NewMessageType(md), nil
}
