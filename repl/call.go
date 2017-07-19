package repl

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func call(env *Env, name string) (string, error) {
	var svc, rpc string
	if env.state.currentService == "" {
		splitted := strings.Split(name, ".")
		if len(splitted) < 2 {
			return "", errors.Wrap(ErrArgumentRequired, "service or RPC name")
		}
		svc, rpc = splitted[0], splitted[1]
	} else {
		svc = env.state.currentService
		rpc = name
	}
	fmt.Println(svc, rpc)
	return "", nil
}
