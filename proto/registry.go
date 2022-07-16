package proto

import (
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

func RegisterFileAndType(fd *desc.FileDescriptor) error {
	deps := fd.GetDependencies()
	for _, d := range deps {
		if err := RegisterFileAndType(d); err != nil {
			return err
		}
	}

	prfd, err := protodesc.NewFile(fd.AsFileDescriptorProto(), protoregistry.GlobalFiles)
	if err != nil {
		return errors.Wrap(err, "failed to new file descriptor")
	}

	if _, err := protoregistry.GlobalFiles.FindFileByPath(prfd.Path()); errors.Is(err, protoregistry.NotFound) {
		if err := protoregistry.GlobalFiles.RegisterFile(prfd); err != nil {
			return err
		}
	}

	for i := 0; i < prfd.Messages().Len(); i++ {
		md := prfd.Messages().Get(i)
		if _, err := protoregistry.GlobalTypes.FindMessageByName(md.FullName()); errors.Is(err, protoregistry.NotFound) {
			if err := protoregistry.GlobalTypes.RegisterMessage(dynamicpb.NewMessageType(md)); err != nil {
				return err
			}
		}
	}

	return nil
}

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
	md, err := r.descSource.FindSymbol(string(m))
	if err != nil {
		return nil, err
	}
	if errors.Is(err, errSymbolNotFound) {
		// Fallback to protoregistry.GlobalTypes.
		return protoregistry.GlobalTypes.FindMessageByName(m)
	}

	return dynamicpb.NewMessageType(md), nil
}

func (r *anyResolver) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	n := strings.LastIndex(url, "/")
	if n != -1 {
		url = url[n+1:]
	}

	md, err := r.descSource.FindSymbol(url)
	if err != nil && !errors.Is(err, errSymbolNotFound) {
		return nil, err
	}
	if errors.Is(err, errSymbolNotFound) {
		// Fallback to protoregistry.GlobalTypes.
		return protoregistry.GlobalTypes.FindMessageByURL(url)
	}

	return dynamicpb.NewMessageType(md), nil
}
