package parser

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/pkg/errors"
)

func ParseFile(filename []string, paths []string) (*FileDescriptorSet, error) {
	args := []string{
		fmt.Sprintln("--proto_path=%s", strings.Join(paths, ":")),
		"--proto_path=.",
		"--include_source_info",
		"--include_imports",
		"--descriptor_set_out=/dev/stdout",
	}

	for _, file := range filename {
		args = append(args, file)
	}

	code, err := runProtoc(args)
	if err != nil {
		return nil, err
	}

	ds := descriptor.FileDescriptorSet{}
	if err := proto.Unmarshal(code, &ds); err != nil {
		return nil, err
	}

	set := make([]*desc.FileDescriptor, len(ds.GetFile()))
	for i, d := range ds.GetFile() {
		var err error
		set[i], err = desc.CreateFileDescriptor(d)
		if err != nil {
			return nil, err
		}
	}

	return &FileDescriptorSet{set}, nil
}

func runProtoc(args []string) ([]byte, error) {
	buf, errBuf := new(bytes.Buffer), new(bytes.Buffer)
	cmd := exec.Command("protoc", args...)
	cmd.Stdout = buf
	cmd.Stderr = errBuf
	if err := cmd.Run(); err != nil {
		if errBuf.Len() != 0 {
			return nil, errors.Wrap(err, errBuf.String())
		}
		return nil, err
	}

	return buf.Bytes(), nil
}
