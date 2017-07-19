package repl

import (
	"strings"

	"github.com/pkg/errors"
)

func describe(env *Env, target string) (string, error) {
	var pack, name string
	splitted := strings.Split(target, ".")
	if env.state.currentPackage == "" {
		if len(splitted) < 2 {
			return "", errors.Wrap(ErrArgumentRequired, "package name")
		}
		pack, name = splitted[0], splitted[1]
	} else {
		pack = env.state.currentPackage
		name = target
	}
	return env.Desc.GetMessage(pack, name).String(), nil
}
