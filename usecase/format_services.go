package usecase

import (
	"sort"

	"github.com/pkg/errors"
)

// FormatServices formats all package names.
func FormatServices() (string, error) {
	return dm.FormatServices()
}
func (m *dependencyManager) FormatServices() (string, error) {
	svcs, err := m.ListServices()
	if err != nil {
		return "", errors.Wrap(err, "failed to list services")
	}
	type service struct {
		Service      string `json:"service"`
		RPC          string `json:"rpc"`
		RequestType  string `json:"request type" table:"request type"`
		ResponseType string `json:"response type" table:"response type"`
	}
	var v struct {
		Services []service `json:"services"`
	}
	for _, svcName := range svcs {
		rpcs, err := m.ListRPCs(svcName)
		if err != nil {
			return "", errors.Wrapf(err, "failed to list RPCs associated with '%s'", svcName)
		}
		for _, rpc := range rpcs {
			v.Services = append(v.Services, service{
				svcName,
				rpc.Name,
				rpc.RequestType.Name,
				rpc.ResponseType.Name,
			})
		}
	}
	sort.Slice(v.Services, func(i, j int) bool {
		return v.Services[i].Service < v.Services[j].Service
	})
	out, err := m.resourcePresenter.Format(v, "  ")
	if err != nil {
		return "", errors.Wrap(err, "failed to format service names by presenter")
	}
	return out, nil
}
