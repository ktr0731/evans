package entity

import (
	"testing"

	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	setup := func(t *testing.T, cfg *config.Env) *Env {
		p, err := toEntitiesFrom([]*desc.FileDescriptor{parseFile(t, "helloworld.proto")})
		require.NoError(t, err)
		env, err := New(p, cfg)
		require.NoError(t, err)
		return env
	}

	env := setup(t, nil)

	t.Run("HasCurrentPackage", func(t *testing.T) {
		assert.False(t, env.HasCurrentPackage())
		env.UsePackage("helloworld")
		assert.True(t, env.HasCurrentPackage())
	})

	t.Run("HasCurrentService", func(t *testing.T) {
		assert.False(t, env.HasCurrentService())
		env.UseService("Greeter")
		assert.True(t, env.HasCurrentService())
	})

	t.Run("Packages", func(t *testing.T) {
		pkgs := env.Packages()
		assert.Len(t, pkgs, 1)
	})

	t.Run("Services", func(t *testing.T) {
		svcs, err := env.Services()
		require.NoError(t, err)
		assert.Len(t, svcs, 1)
	})

	t.Run("Messages", func(t *testing.T) {
		msgs, err := env.Messages()
		require.NoError(t, err)
		assert.Len(t, msgs, 2)
	})

	t.Run("RPCs", func(t *testing.T) {
		rpcs, err := env.RPCs()
		require.NoError(t, err)
		assert.Len(t, rpcs, 1)
	})

	t.Run("Service", func(t *testing.T) {
		svc, err := env.Service("Greeter")
		require.NoError(t, err)
		assert.Equal(t, "Greeter", svc.Name)
		assert.Len(t, svc.RPCs, 1)
	})

	t.Run("Message", func(t *testing.T) {
		msg, err := env.Message("HelloRequest")
		require.NoError(t, err)
		assert.Equal(t, "HelloRequest", msg.Name())
		assert.Len(t, msg.Fields(), 2)
	})

	t.Run("RPC", func(t *testing.T) {
		rpc, err := env.RPC("SayHello")
		require.NoError(t, err)
		assert.Equal(t, "SayHello", rpc.Name)
	})

	t.Run("Headers", func(t *testing.T) {
		cfg := &config.Env{
			Request: &config.Request{},
		}
		env := setup(t, cfg)
		assert.Len(t, env.Headers(), 0)

		expected := []config.Header{
			{Key: "foo", Val: "bar"},
			{Key: "hoge", Val: "fuga"},
		}
		cfg.Request.Header = expected
		env = setup(t, cfg)
		for i, h := range env.Headers() {
			assert.Equal(t, expected[i].Key, h.Key)
			assert.Equal(t, expected[i].Val, h.Val)
		}
	})

	t.Run("AddHeader", func(t *testing.T) {
		cfg := &config.Env{
			Request: &config.Request{},
		}
		env := setup(t, cfg)
		require.Len(t, env.Headers(), 0)

		err := env.AddHeader(&Header{"megumi", "kato", false})
		require.NoError(t, err)
		assert.Len(t, env.Headers(), 1)

		err = env.AddHeader(&Header{"megumi", "kato", false})
		require.Error(t, err)
	})

	t.Run("RemoveHeader", func(t *testing.T) {
		cfg := &config.Env{
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
		assert.Len(t, env.Headers(), 4)

		env.RemoveHeader("foo")
		assert.Len(t, env.Headers(), 4)

		env.RemoveHeader("sapphire")
		assert.Len(t, env.Headers(), 3)
		assert.Equal(t, env.Headers()[2].Key, "hazuki")

		env.RemoveHeader("hazuki")
		assert.Len(t, env.Headers(), 2)
		assert.Equal(t, env.Headers()[1].Key, "reina")
	})
}
