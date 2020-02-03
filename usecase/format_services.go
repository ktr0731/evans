package usecase

import (
	"sort"

	"github.com/ktr0731/evans/idl"
	"github.com/pkg/errors"
)

// FormatServicesParams is the modifier for FormatServices.
type FormatServicesParams struct {
	// FullyQualifiedName means format with fully-qualified service names if it is true.
	FullyQualifiedName bool
}

// FormatServices formats all service names the spec loaded.
func FormatServices(p *FormatServicesParams) (string, error) {
	return dm.FormatServices(p)
}
func (m *dependencyManager) FormatServices(p *FormatServicesParams) (string, error) {
	if p == nil {
		p = &FormatServicesParams{}
	}

	svcs, err := m.ListServices()
	if err != nil {
		return "", errors.Wrap(err, "failed to list services")
	}
	type svc struct {
		Service string `json:"service"`
	}
	var v struct {
		Services []svc `json:"services"`
	}
	for _, svcName := range svcs {
		if p.FullyQualifiedName {
			svcName, err = idl.FullyQualifiedServiceName(m.state.selectedPackage, svcName)
			if err != nil {
				return "", errors.Wrap(err, "failed to get fully-qualified service name")
			}
		}
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
