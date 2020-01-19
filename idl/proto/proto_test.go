package proto_test

import (
	"errors"
	"testing"

	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/grpc/grpcreflection"
	"github.com/ktr0731/evans/idl/proto"
)

func TestLoadFiles(t *testing.T) {
	cases := map[string]struct {
		fnames []string
		hasErr bool
	}{
		"normal":        {fnames: []string{"message.proto", "api.proto", "other_package.proto"}},
		"invalid proto": {fnames: []string{"invalid.proto"}, hasErr: true},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			_, err := proto.LoadFiles([]string{"testdata"}, c.fnames)
			if c.hasErr {
				if err == nil {
					t.Errorf("LoadFiles must return an error, but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("LoadFiles must not return an error, but got '%s'", err)
			}
		})
	}
}

type reflectionClient struct {
	grpcreflection.Client
	descs []*desc.FileDescriptor
	err   error
}

func (c *reflectionClient) ListPackages() ([]*desc.FileDescriptor, error) {
	return c.descs, c.err
}

func TestLoadByReflection(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		refCli := &reflectionClient{}
		_, err := proto.LoadByReflection(refCli)
		if err != nil {
			t.Errorf("must not return an error, but got '%s'", err)
		}
	})

	t.Run("reflection client returns an error", func(t *testing.T) {
		refCli := &reflectionClient{err: errors.New("an err")}
		_, err := proto.LoadByReflection(refCli)
		if err == nil {
			t.Errorf("must return an error, but got nil")
		}
	})
}
