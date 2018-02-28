package parser

import (
	"github.com/ktr0731/evans/adapter/internal/protoparser"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/entity"
)

func ParseFile(filename []string, paths []string) ([]*entity.Package, error) {
	set, err := protoparser.ParseFile(filename, paths)
	if err != nil {
		return nil, err
	}
	return protobuf.ToEntitiesFrom(set)
}
