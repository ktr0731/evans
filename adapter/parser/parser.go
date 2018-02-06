package parser

import (
	"github.com/ktr0731/evans/adapter/internal/proto_parser"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/entity"
)

func ParseFile(filename []string, paths []string) (entity.Packages, error) {
	set, err := proto_parser.ParseFile(filename, paths)
	if err != nil {
		return nil, err
	}
	return protobuf.ToEntitiesFrom(set)
}
