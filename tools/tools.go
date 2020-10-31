// +build tools

package tools

import (
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/kisielk/godepgraph"
	_ "github.com/ktr0731/bump"
	_ "github.com/matryer/moq"
)
