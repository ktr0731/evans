package entity

import (
	"testing"

	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/parser"
	"github.com/stretchr/testify/require"
)

func fileDesc(t *testing.T, name, path []string) *parser.FileDescriptorSet {
	desc, err := parser.ParseFile(name, path)
	require.NoError(t, err)
	return desc
}

func getMessage(t *testing.T, fd *parser.FileDescriptorSet, pkgName, msgName string) *Message {
	for _, m := range fd.GetMessages(pkgName) {
		if m.GetName() == msgName {
			return NewMessage(m)
		}
	}
	t.Fatal("message not found")
	return nil
}

func getService(t *testing.T, fd *parser.FileDescriptorSet, pkgName, svcName string) *desc.ServiceDescriptor {
	for _, s := range fd.GetServices(pkgName) {
		if s.GetName() == svcName {
			return s
		}
	}
	t.Fatal("service not found")
	return nil
}
