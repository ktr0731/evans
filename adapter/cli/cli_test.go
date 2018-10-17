package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCommandLineMode(t *testing.T) {
	cases := map[string]struct {
		stdin             io.Reader
		file              string
		isCommandLineMode bool
	}{
		"true because stdin is provided":                      {stdin: &bytes.Buffer{}, isCommandLineMode: true},
		"true because file is passed":                         {file: "touma.json", isCommandLineMode: true},
		"true because both of stdin and file are passed":      {stdin: &bytes.Buffer{}, file: "touma.json", isCommandLineMode: true},
		"false because both of stdin and file are not passed": {},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.stdin != nil {
				cleanup := func() func() {
					stdin := os.Stdin
					return func() {
						os.Stdin = stdin
					}
				}()
				defer cleanup()

				result := IsCommandLineMode(c.file)
				assert.Equal(t, c.isCommandLineMode, result)
			}
		})
	}
}
