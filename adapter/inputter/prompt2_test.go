package inputter

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jhump/protoreflect/desc/protoparse"
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
