package model

import (
	"testing"

	"github.com/lycoris0731/evans/parser"
	"github.com/stretchr/testify/require"
)

func fileDesc(t *testing.T, name, path []string) *parser.FileDescriptorSet {
	desc, err := parser.ParseFile(name, path)
	require.NoError(t, err)
	return desc
}

func getMessage(t *testing.T, desc *parser.FileDescriptorSet, pkgName, msgName string) *Message {
	for _, m := range desc.GetMessages(pkgName) {
		if m.GetName() == msgName {
			return NewMessage(m)
		}
	}
	t.Fatal("message not found")
	return nil
}
