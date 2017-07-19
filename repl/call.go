package repl

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func call(state *State, name string) (string, error) {
	var svc, rpc string
	if state.currentService == "" {
		splitted := strings.Split(name, ".")
		if len(splitted) < 2 {
			return "", errors.Wrap(ErrArgumentRequired, "service or RPC name")
		}
		svc, rpc = splitted[0], splitted[1]
	} else {
		svc = state.currentService
		rpc = name
	}
	fmt.Println(svc, rpc)
	return "", nil
}
