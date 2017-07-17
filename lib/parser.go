package lib

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/pkg/errors"
)

func ParseFile(filename string, paths ...string) (*FileDescriptorSet, error) {
	args := []string{
		fmt.Sprintln("--proto_path=%s", strings.Join(paths, ":")),
		"--proto_path=.",
		"--include_source_info",
		"--include_imports",
		"--descriptor_set_out=/dev/stdout",
		filename,
	}

	code, err := runProtoc(args)
	if err != nil {
		return nil, err
	}

	desc := descriptor.FileDescriptorSet{}
	if err := proto.Unmarshal(code, &desc); err != nil {
		return nil, err
	}

	return &FileDescriptorSet{&desc}, nil
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
