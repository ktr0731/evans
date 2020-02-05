package usecase

import (
	"sort"

	"github.com/pkg/errors"
)

// FormatRPCs formats all package names.
func FormatRPCs() (string, error) {
	return dm.FormatRPCs()
}
func (m *dependencyManager) FormatRPCs() (string, error) {
	svcs := m.ListServices()
	type rpc struct {
		RPC          string `json:"rpc"`
		RequestType  string `json:"request type" table:"request type"`
		ResponseType string `json:"response type" table:"response type"`
	}
	var v struct {
		RPCs []rpc `json:"rpcs"`
	}
	for _, svcName := range svcs {
		rpcs, err := m.ListRPCs(svcName)
		if err != nil {
			return "", errors.Wrapf(err, "failed to list RPCs associated with '%s'", svcName)
		}
		for _, r := range rpcs {
			v.RPCs = append(v.RPCs, rpc{
				r.Name,
				r.RequestType.Name,
				r.ResponseType.Name,
			})
		}
	}
	sort.Slice(v.RPCs, func(i, j int) bool {
		return v.RPCs[i].RPC < v.RPCs[j].RPC
	})
	out, err := m.resourcePresenter.Format(v, "  ")
	if err != nil {
		return "", errors.Wrap(err, "failed to format RPC names by presenter")
	}
	return out, nil
}
