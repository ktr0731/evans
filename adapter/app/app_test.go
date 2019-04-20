package app

import (
	"testing"

	"github.com/ktr0731/evans/adapter/cui"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestNew(t *testing.T) {
	cases := map[string]struct {
		ui cui.UI
	}{
		"normal": {ui: cui.New()},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			cmd := New(c.ui)
			assert.NotNil(t, cmd.ui)
		})
	}
}
