package repl

import (
	"github.com/pkg/errors"
)

func show(env *Env, target string) (string, error) {
	switch target {
	case "s", "svc", "service", "services":
		return env.Desc.GetServices().String(), nil

	case "p", "package", "packages":
		return env.Desc.GetPackages().String(), nil

	case "m", "message", "messages":
		return env.Desc.GetMessages().String(), nil

	default:
		return "", errors.Wrap(ErrUnknownTarget, target)
	}
}
