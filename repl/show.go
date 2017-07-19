package repl

import (
	"github.com/lycoris0731/evans/env"
	"github.com/pkg/errors"
)

func show(env *env.Env, target string) (string, error) {
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
		msg, err := env.GetMessages()
		if err != nil {
			return "", err
		}
		return msg.String(), nil

	default:
		return "", errors.Wrap(ErrUnknownTarget, target)
	}
}
