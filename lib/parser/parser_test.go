package parser

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func Test_runProtoc(t *testing.T) {
	tests := []struct {
		targets     []string
		expectedErr error
	}{
		{
			targets:     []string{"../../testdata/proto/test.proto"},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		paths := make([]string, len(test.targets))
		for i, target := range test.targets {
			paths[i] = filepath.Dir(target)
		}

		args := []string{
			fmt.Sprintf("--proto_path=%s", strings.Join(paths, ":")),
			"--include_source_info",
			"--include_imports",
			"--descriptor_set_out=/dev/stdout",
		}
		args = append(args, test.targets...)

		code, err := runProtoc(args)
		if err != test.expectedErr {
			t.Fatalf("expectedErr: %s, actual: %s", test.expectedErr, err)
		}

		if len(code) == 0 {
			t.Fatalf("returned byte length is 0")
		}
	}
}
