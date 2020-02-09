package usecase

import (
	"sort"

	"github.com/pkg/errors"
)

// FormatServicesOld formats all package names.
// Deprecated: dropped in the next major release.
func FormatServicesOld() (string, error) {
	return dm.FormatServicesOld()
}
func (m *dependencyManager) FormatServicesOld() (string, error) {
	svcs := m.ListServicesOld()
	type service struct {
		Service      string `json:"service"`
		RPC          string `json:"rpc"`
		RequestType  string `json:"request type" table:"request type"`
		ResponseType string `json:"response type" table:"response type"`
	}
	var v struct {
		Services []service `json:"services"`
	}
	for _, svc := range svcs {
		rpcs, err := m.ListRPCs(svc)
		if err != nil {
			return "", errors.Wrapf(err, "failed to list RPCs associated with '%s'", svc)
		}
		for _, rpc := range rpcs {
			v.Services = append(v.Services, service{
				svc,
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
