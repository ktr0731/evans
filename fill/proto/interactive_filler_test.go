package proto_test

import (
	"testing"

	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/fill/proto"
)

func TestInteractiveProtoFiller(t *testing.T) {
	f := proto.NewInteractiveFiller(nil, "")
	err := f.Fill("invalid type")
	if err != fill.ErrCodecMismatch {
		t.Errorf("filler must return fil.ErrCodecMismatch because the arg is invalid type, but got unexpected error: %s", err)
	}
}
