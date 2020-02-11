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
	fqsns := m.ListServices()
	type svc struct {
		Name string `json:"name" name:"target"`
	}
	var v struct {
		Services []svc `json:"services" name:"target"`
	}
	for _, fqsn := range fqsns {
		v.Services = append(v.Services, svc{fqsn})
	}
	sort.Slice(v.Services, func(i, j int) bool {
		return v.Services[i].Name < v.Services[j].Name
	})
	out, err := m.resourcePresenter.Format(v)
	if err != nil {
		return "", errors.Wrap(err, "failed to format service names by presenter")
	}
	return out, nil
}
