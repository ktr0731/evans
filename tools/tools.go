// +build tools

package tools

import (
	_ "github.com/Songmu/gocredits/cmd/gocredits"
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/kisielk/godepgraph"
	_ "github.com/ktr0731/bump"
	_ "github.com/matryer/moq"
)
