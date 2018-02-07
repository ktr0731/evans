package entity

import (
	"testing"

	"github.com/ktr0731/evans/config"
	"github.com/stretchr/testify/require"
)

func TestNewEnv(t *testing.T) {
	cfg := &config.Config{
		Request: &config.Request{
			Header: []config.Header{{"foo", "bar"}},
		},
	}
	env, err := NewEnv(nil, cfg)
	require.NoError(t, err)
	h := env.Headers()
	require.Len(t, h, 1)
	require.Equal(t, h[0].Key, "foo")
	require.Equal(t, h[0].Val, "bar")
}

func TestEnv(t *testing.T) {
	pkgs := []*Package{
		{
			Name: "helloworld",
			Services: []Service{
				&svc{
					name: "Greeter",
					rpcs: []*rpc{
						&rpc{name: "SayHello"},
					},
				},
			},
			Messages: []Message{
				&msg{
					name: "HelloRequest",
					fields: []Field{
						&fld{name: "name"},
					},
				},
				&msg{name: "HelloResponse"},
			},
		},
	}
	setup := func(t *testing.T, cfg *config.Config) *Env {
		if cfg == nil {
			cfg = &config.Config{
				Env: &config.Env{},
				Request: &config.Request{
					Header: []config.Header{},
				},
			}
		}
		env, err := NewEnv(pkgs, cfg)
		require.NoError(t, err)
		return env
	}

	env := setup(t, nil)

	t.Run("HasCurrentPackage", func(t *testing.T) {
		require.False(t, env.HasCurrentPackage())
		env.UsePackage("helloworld")
		require.True(t, env.HasCurrentPackage())
	})

	t.Run("HasCurrentService", func(t *testing.T) {
		require.False(t, env.HasCurrentService())
		env.UseService("Greeter")
		require.True(t, env.HasCurrentService())
	})

	t.Run("Packages", func(t *testing.T) {
		pkgs := env.Packages()
		require.Len(t, pkgs, 1)
	})

	t.Run("Services", func(t *testing.T) {
		svcs, err := env.Services()
		require.NoError(t, err)
		require.Len(t, svcs, 1)
	})

	t.Run("Messages", func(t *testing.T) {
		msgs, err := env.Messages()
		require.NoError(t, err)
		require.Len(t, msgs, 2)
	})

	t.Run("RPCs", func(t *testing.T) {
		rpcs, err := env.RPCs()
		require.NoError(t, err)
		require.Len(t, rpcs, 1)
	})

	t.Run("Service", func(t *testing.T) {
		svc, err := env.Service("Greeter")
		require.NoError(t, err)
		require.Equal(t, "Greeter", svc.Name())
		require.Len(t, svc.RPCs(), 1)
	})

	t.Run("Message", func(t *testing.T) {
		msg, err := env.Message("HelloRequest")
		require.NoError(t, err)
		require.Equal(t, "HelloRequest", msg.Name())
		require.Len(t, msg.Fields(), 1)
	})

	t.Run("RPC", func(t *testing.T) {
		rpc, err := env.RPC("SayHello")
		require.NoError(t, err)
		require.Equal(t, "SayHello", rpc.Name())
	})

	t.Run("Headers", func(t *testing.T) {
		cfg := &config.Config{
			Request: &config.Request{},
		}
		env := setup(t, cfg)
		require.Len(t, env.Headers(), 0)

		expected := []config.Header{
			{Key: "foo", Val: "bar"},
			{Key: "hoge", Val: "fuga"},
		}
		cfg.Request.Header = expected
		env = setup(t, cfg)
		for i, h := range env.Headers() {
			require.Equal(t, expected[i].Key, h.Key)
			require.Equal(t, expected[i].Val, h.Val)
		}
	})

	t.Run("AddHeader", func(t *testing.T) {
		cfg := &config.Config{
			Request: &config.Request{},
		}
		env := setup(t, cfg)
		require.Len(t, env.Headers(), 0)

		err := env.AddHeader(&Header{"megumi", "kato", false})
		require.NoError(t, err)
		require.Len(t, env.Headers(), 1)

		err = env.AddHeader(&Header{"megumi", "kato", false})
		require.Error(t, err)
	})

	t.Run("RemoveHeader", func(t *testing.T) {
		cfg := &config.Config{
			Request: &config.Request{},
		}
		env := setup(t, cfg)
		require.Len(t, env.Headers(), 0)

		env.RemoveHeader("foo")

		headers := []struct {
			k, v string
		}{
			{"kumiko", "oumae"},
			{"reina", "kousaka"},
			{"sapphire", "kawashima"},
			{"hazuki", "katou"},
		}
		for _, h := range headers {
			err := env.AddHeader(&Header{h.k, h.v, false})
			require.NoError(t, err)
		}
		require.Len(t, env.Headers(), 4)

		env.RemoveHeader("foo")
		require.Len(t, env.Headers(), 4)

		env.RemoveHeader("sapphire")
		require.Len(t, env.Headers(), 3)
		require.Equal(t, env.Headers()[2].Key, "hazuki")

		env.RemoveHeader("hazuki")
		require.Len(t, env.Headers(), 2)
		require.Equal(t, env.Headers()[1].Key, "reina")
	})
}

// stubs

type fld struct {
	Field

	name string
}

func (f *fld) Name() string {
	return f.name
}

type rpc struct {
	RPC

	name string
}

func (r *rpc) Name() string {
	return r.name
}

type svc struct {
	Service

	name string
	rpcs []*rpc
}

func (s *svc) Name() string {
	return s.name
}

func (s *svc) RPCs() []RPC {
	rpcs := make([]RPC, 0, len(s.rpcs))
	for _, rpc := range s.rpcs {
		rpcs = append(rpcs, rpc)
	}
	return rpcs
}

type msg struct {
	Message

	name   string
	fields []Field
}

func (m *msg) Name() string {
	return m.name
}

func (m *msg) Fields() []Field {
	return m.fields
}
