package protobuf

import (
	"path/filepath"
	"testing"

	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/adapter/internal/protoparser"
	"github.com/ktr0731/evans/entity"
	"github.com/stretchr/testify/require"
)

func parseFile(t *testing.T, fnames []string, paths []string) []*desc.FileDescriptor {
	for i := range fnames {
		fnames[i] = filepath.Join("testdata", fnames[i])
	}
	d, err := protoparser.ParseFile(fnames, paths)
	require.NoError(t, err)
	return d
}

func toEntitiesFrom(files []*desc.FileDescriptor) ([]*entity.Package, error) {
	var pkgNames []string
	msgMap := map[string][]entity.Message{}
	svcMap := map[string][]entity.Service{}
	for _, f := range files {
		pkgName := f.GetPackage()

		pkgNames = append(pkgNames, pkgName)

		for _, msg := range f.GetMessageTypes() {
			msgMap[pkgName] = append(msgMap[pkgName], newMessage(msg))
		}
		for _, svc := range f.GetServices() {
			svcMap[pkgName] = append(svcMap[pkgName], newService(svc))
		}
	}

	var pkgs []*entity.Package
	for _, pkgName := range pkgNames {
		pkgs = append(pkgs, entity.NewPackage(pkgName, msgMap[pkgName], svcMap[pkgName]))
	}

	return pkgs, nil
}

func testdata(s ...string) string {
	return filepath.Join(append([]string{"testdata"}, s...)...)
}
