package inputter

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/ktr0731/evans/adapter/prompt"
	"github.com/stretchr/testify/require"
)

func TestPromptInputter2(t *testing.T) {
	p := &protoparse.Parser{}
	fd, err := p.ParseFiles(filepath.Join("testdata", "oneof.proto"))
	require.NoError(t, err)
	for _, msg := range fd[0].GetMessageTypes() {
		fmt.Printf("-- %s --\n", msg.GetName())
		NewPromptV2("", nil).Input(msg)
		fmt.Println()
	}
}

func TestFoo(t *testing.T) {
	p := &protoparse.Parser{}
	fd, err := p.ParseFiles(filepath.Join("testdata", "oneof.proto"))
	require.NoError(t, err)
	pr := NewPromptV2("", nil)
	pr.prompt = prompt.New(func(s string) {
		fmt.Println(s)
	}, nil)
	pr.Input(fd[0].GetMessageTypes()[0])
}
