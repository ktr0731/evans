package parser

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
	"github.com/pkg/errors"
)

func ParseFile(filename []string, paths []string) (entity.Packages, error) {
	args := []string{
		fmt.Sprintf("--proto_path=%s", strings.Join(paths, ":")),
		"--proto_path=.",
		"--include_source_info",
		"--include_imports",
		"--descriptor_set_out=/dev/stdout",
	}

	args = append(args, filename...)

	code, err := runProtoc(args)
	if err != nil {
		return nil, err
	}

	ds := descriptor.FileDescriptorSet{}
	if err := proto.Unmarshal(code, &ds); err != nil {
		return nil, err
	}

	set := make([]*desc.FileDescriptor, len(ds.GetFile()))
	files := ds.GetFile()
	// sort files by number of dependencies
	sort.Slice(files, func(i, j int) bool {
		return len(files[i].GetDependency()) < len(files[j].GetDependency())
	})

	depsCache := map[string]*desc.FileDescriptor{}
	for i, d := range files {
		var err error

		// collect dependencies
		deps := make([]*desc.FileDescriptor, len(d.GetDependency()))
		for i, depName := range d.GetDependency() {
			deps[i] = depsCache[depName]
		}

		set[i], err = desc.CreateFileDescriptor(d, deps...)
		if err != nil {
			return nil, err
		}
		depsCache[d.GetName()] = set[i]
	}

	return toEntitiesFrom(set)
}

// toEntitiesFrom normalizes descriptors to entities
//
// package
// ├ messages
// ├ enums
// └ services
//   └ rpcs
//
func toEntitiesFrom(files []*desc.FileDescriptor) (entity.Packages, error) {
	// TODO: file -> packages へ正規化する
	var pkgNames []string
	msgMap := map[string][]*entity.Message{}
	svcMap := map[string][]*entity.Service{}
	for _, f := range files {
		pkgName := f.GetPackage()

		pkgNames = append(pkgNames, pkgName)

		for _, msg := range f.GetMessageTypes() {
			msgMap[pkgName] = append(msgMap[pkgName], entity.NewMessage(msg))
		}
		for _, svc := range f.GetServices() {
			svcMap[pkgName] = append(svcMap[pkgName], entity.NewService(svc))
		}
	}

	var pkgs entity.Packages
	for _, pkgName := range pkgNames {
		pkgs = append(pkgs, entity.NewPackage(pkgName, msgMap[pkgName], svcMap[pkgName]))
	}

	return pkgs, nil
}

func runProtoc(args []string) ([]byte, error) {
	buf, errBuf := new(bytes.Buffer), new(bytes.Buffer)
	cmd := exec.Command("protoc", args...)
	cmd.Stdout = buf
	cmd.Stderr = errBuf
	if err := cmd.Run(); err != nil {
		if errBuf.Len() != 0 {
			return nil, errors.New(errBuf.String())
		}
		return nil, err
	}

	return buf.Bytes(), nil
}
