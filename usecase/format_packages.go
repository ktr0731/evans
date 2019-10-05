package usecase

import (
	"sort"

	"github.com/pkg/errors"
)

// FormatPackages formats all package names.
func FormatPackages() (string, error) {
	return dm.FormatPackages()
}
func (m *dependencyManager) FormatPackages() (string, error) {
	pkgs := m.ListPackages()
	type pkg struct {
		Package string `json:"package"`
	}
	var v struct {
		Packages []pkg `json:"packages"`
	}
	for _, pkgName := range pkgs {
		v.Packages = append(v.Packages, pkg{pkgName})
	}
	sort.Slice(v.Packages, func(i, j int) bool {
		return v.Packages[i].Package < v.Packages[j].Package
	})
	out, err := m.resourcePresenter.Format(v, "  ")
	if err != nil {
		return "", errors.Wrap(err, "failed to format package names by presenter")
	}
	return out, nil
}
