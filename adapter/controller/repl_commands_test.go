package controller

import (
	"io"
	"testing"

	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/require"
)

type headInputPort struct {
	port.InputPort
	received *port.HeaderParams
}

func (i *headInputPort) Header(p *port.HeaderParams) (io.Reader, error) {
	i.received = p
	return nil, nil
}

func Test_headCommand(t *testing.T) {
	cases := map[string]struct {
		in       string
		expected *entity.Header
		hasError bool
	}{
		"normal":                      {in: "tanamachi=kaoru", expected: &entity.Header{Key: "tanamachi", Val: "kaoru"}},
		"normal (has a double-quote)": {in: `nanasaki=ai"`, expected: &entity.Header{Key: "nanasaki", Val: `ai"`}},
		"normal (has a = in value)":   {in: "morishima=lovely=haruka", expected: &entity.Header{Key: "morishima", Val: "lovely=haruka"}},
		"remove key1":                 {in: "amagami-_.0", expected: &entity.Header{Key: "amagami-_.0", NeedToRemove: true}},
		"remove key2":                 {in: "rihoko=", expected: &entity.Header{Key: "rihoko", Val: "", NeedToRemove: true}},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			inputPort := &headInputPort{}
			cmd := &headerCommand{inputPort}

			_, err := cmd.Run([]string{c.in})
			if c.hasError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Exactly(t, c.expected, inputPort.received.Headers[0])
		})
	}
}
