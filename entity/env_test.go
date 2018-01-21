package entity

import (
	"testing"

	"github.com/jhump/protoreflect/desc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	setup := func(t *testing.T) *Env {
		p, err := toEntitiesFrom([]*desc.FileDescriptor{parseFile(t, "helloworld.proto")})
		require.NoError(t, err)
		env, err := New(p, nil)
		require.NoError(t, err)
		return env
	}

	env := setup(t)

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
		assert.Len(t, msg.Fields, 2)
	})

	t.Run("RPC", func(t *testing.T) {
		rpc, err := env.RPC("SayHello")
		require.NoError(t, err)
		assert.Equal(t, "SayHello", rpc.Name)
	})
}
