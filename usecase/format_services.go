package usecase

import (
	"sort"

	"github.com/pkg/errors"
)

// FormatServices formats all service names the spec loaded.
func FormatServices() (string, error) {
	return dm.FormatServices()
}
func (m *dependencyManager) FormatServices() (string, error) {
	svcs := m.ListServices()
	type svc struct {
		Service string `json:"service"`
	}
	var v struct {
		Services []svc `json:"services"`
	}
	for _, svcName := range svcs {
		v.Services = append(v.Services, svc{svcName})
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
