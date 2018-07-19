package config

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestConfig_Get(t *testing.T) {
	checkValues := func() {
		c := Get()
		_ = c.REPL.Server.Host
		_ = c.Env.Server.Host
	}

	t.Run("no local config", func(t *testing.T) {
		checkValues()
	})

	t.Run("has local config", func(t *testing.T) {
		f, err := ioutil.TempFile("", "")
		require.NoError(t, err)
		defer f.Close()

		backup := localConfigName
		defer func() {
			localConfigName = backup
		}()

		localConfigName = f.Name()

		checkValues()
	})

	t.Run("REPL.Server and Env.Server are same as top-level Server", func(t *testing.T) {
		checkValues()
		c := Get()
		assert.Equal(t, c.Server, c.Env.Server)
		assert.Equal(t, c.Server, c.REPL.Server)
	})
}
