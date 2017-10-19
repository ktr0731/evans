package repl

import (
	"strings"

	"github.com/ktr0731/evans/env"
	"github.com/pkg/errors"
)

func show(env *env.Env, target string) (string, error) {
	switch strings.ToLower(target) {
	case "p", "package", "packages":
		return env.GetPackages().String(), nil

	case "s", "svc", "service", "services":
		svc, err := env.GetServices()
		if err != nil {
			return "", err
		}
		return svc.String(), nil

	case "m", "msg", "message", "messages":
		msg, err := env.GetMessages()
		if err != nil {
			return "", err
		}
		return msg.String(), nil

	case "a", "r", "rpc", "api":
		rpcs, err := env.GetRPCs()
		if err != nil {
			return "", err
		}
		return rpcs.String(), nil

	default:
		return "", errors.Wrap(ErrUnknownTarget, target)
	}
}
