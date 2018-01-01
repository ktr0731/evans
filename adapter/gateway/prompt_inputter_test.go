package gateway

import (
	"os"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/require"
)

func TestPromptInputter_Input(t *testing.T) {
	pp.Println(os.Getwd())
	set := helper.ReadProto(t, []string{"helloworld/helloworld.proto"})
	env := helper.NewEnv(t, set, helper.TestConfig().Env)
	NewPromptInputter(env)

	err := env.UsePackage("helloworld")
	require.NoError(t, err)

	// inputter.Input()
}
