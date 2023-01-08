package usecase

import (
	"sort"

	"github.com/ktr0731/evans/proto"
)

// ListPackages lists all package names.
func ListPackages() []string {
	return dm.ListPackages()
}
func (m *dependencyManager) ListPackages() []string {
	pkgMap := map[string]struct{}{}
	for _, s := range m.descSource.ListServices() {
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

	return pkgs
}
