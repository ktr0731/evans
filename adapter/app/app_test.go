package app

import (
	"testing"

	"github.com/ktr0731/evans/adapter/cui"
	"github.com/ktr0731/evans/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func Test_mergeConfig(t *testing.T) {
	// setup
	// append elements to slice which will be merged.
	cfg := &config.Config{
		Default: &config.Default{
			Package:   "tamaki",
			ProtoPath: []string{"kobuchizawa"},
			ProtoFile: []string{"miyake"},
		},
		Server: &config.Server{
			Port: "50052",
		},
		Request: &config.Request{
			Header: []config.Header{
				{Key: "yuzuki", Val: "shiraishi"},
				{Key: "nozomi", Val: "kasaki"},
				{Key: "nozomi", Val: "kasaki2"},
			},
		},
		Env:  &config.Env{},
		REPL: &config.REPL{},
	}
	config.SetupConfig(cfg)

	opt := &options{
		pkg:     "kumiko",
		service: "reina",
		path:    []string{"kobuchizawa", "midori"},
		host:    "hazuki",
		port:    "50053",
		header:  []string{"nozomi=kasaki"},
	}
	proto := []string{"noboru"}

	res, err := mergeConfig(cfg, opt, proto)
	require.NoError(t, err)
	assert.Equal(t, res.REPL.Server, res.Server)
	assert.Equal(t, res.Env.Server, res.Server)
	assert.Equal(t, opt.pkg, res.Default.Package)
	assert.Equal(t, opt.service, res.Default.Service)
	assert.Equal(t, []string(opt.path), res.Default.ProtoPath)
	assert.Equal(t, append(cfg.Default.ProtoFile, proto...), res.Default.ProtoFile)
	assert.Equal(t, opt.host, res.Server.Host)
	assert.Equal(t, opt.port, res.Server.Port)
	assert.Equal(t, cfg.Request.Header[:2], res.Request.Header)
}

func TestNew(t *testing.T) {
	cases := map[string]struct {
		ui cui.UI
	}{
		"normal": {ui: cui.NewBasic()},
		"if ui is nil, default implementation will be used": {},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			cmd := New(c.ui)
			assert.NotNil(t, cmd.ui)
		})
	}
}
