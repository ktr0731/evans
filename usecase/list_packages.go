package usecase

import (
	"sort"

	"github.com/ktr0731/evans/proto"
)

// ListPackages lists all package names.
func ListPackages() ([]string, error) {
	return dm.ListPackages()
}
func (m *dependencyManager) ListPackages() ([]string, error) {
	pkgMap := map[string]struct{}{}
	svcs, err := m.descSource.ListServices()
	if err != nil {
		return nil, err
	}

	for _, s := range svcs {
		pkg, _ := proto.ParseFullyQualifiedServiceName(s)
		pkgMap[pkg] = struct{}{}
	}

	pkgs := make([]string, 0, len(pkgMap))
	for pkg := range pkgMap {
		pkgs = append(pkgs, pkg)
	}

	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i] < pkgs[j]
	})

	return pkgs, nil
}
