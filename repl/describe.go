package repl

import (
	"github.com/ktr0731/evans/env"
)

func describe(env *env.Env, msgName string) (string, error) {
	msg, err := env.GetMessage(msgName)
	if err != nil {
		return "", err
	}
	return msg.String(), nil
}
