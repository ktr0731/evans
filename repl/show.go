package repl

import (
	"github.com/pkg/errors"
)

func show(env *Env, target string) (string, error) {
	switch target {
	case "p", "package", "packages":
		return env.GetPackages().String(), nil

	case "s", "svc", "service", "services":
		svc, err := env.GetServices()
		if err != nil {
			return "", err
		}
		return svc.String(), nil

	case "m", "message", "messages":
		return env.Desc.GetMessages().String(), nil

	default:
		return "", errors.Wrap(ErrUnknownTarget, target)
	}
}
