package controller

import (
	"io"
	"testing"

	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/assert"
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
		In       string
		Expected *entity.Header
		HasError bool
	}{
		"normal":                      {In: "tanamachi=kaoru", Expected: &entity.Header{Key: "tanamachi", Val: "kaoru"}},
		"normal (has a double-quote)": {In: `nanasaki=ai"`, Expected: &entity.Header{Key: "nanasaki", Val: `ai"`}},
		"normal (has a = in value)":   {In: "morishima=lovely=haruka", Expected: &entity.Header{Key: "morishima", Val: "lovely=haruka"}},
		"remove key1":                 {In: "amagami-_.0", Expected: &entity.Header{Key: "amagami-_.0", NeedToRemove: true}},
		"remove key2":                 {In: "rihoko=", Expected: &entity.Header{Key: "rihoko", Val: "", NeedToRemove: true}},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			inputPort := &headInputPort{}
			cmd := &headerCommand{inputPort}

			_, err := cmd.Run([]string{c.In})
			if c.HasError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Exactly(t, c.Expected, inputPort.received.Headers[0])
		})
	}
}
