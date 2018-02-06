package protobuf

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/stretchr/testify/require"
)

// for prevent cycle import, don't use adapter/parser
func parseFile(t *testing.T, fname string, paths ...string) *desc.FileDescriptor {
	return parseDependFiles(t, fname, paths...)[0]
}

// parseDependFiles is used to marshal importing proto file
func parseDependFiles(t *testing.T, fname string, paths ...string) []*desc.FileDescriptor {
	args := []string{
		fmt.Sprintf("--proto_path=%s", strings.Join(paths, ":")),
		"--proto_path=testdata",
		"--include_source_info",
		"--include_imports",
		"--descriptor_set_out=/dev/stdout",
		filepath.Join("testdata", fname),
	}

	code, err := runProtoc(args)
	require.NoError(t, err)

	ds := descriptor.FileDescriptorSet{}
	err = proto.Unmarshal(code, &ds)
	require.NoError(t, err)

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
		require.NoError(t, err)

		depsCache[d.GetName()] = set[i]
	}

	return set
}

func toEntitiesFrom(files []*desc.FileDescriptor) (Packages, error) {
	var pkgNames []string
	msgMap := map[string][]*Message{}
	svcMap := map[string][]*Service{}
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

	var pkgs Packages
	for _, pkgName := range pkgNames {
		pkgs = append(pkgs, newPackage(pkgName, msgMap[pkgName], svcMap[pkgName]))
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
