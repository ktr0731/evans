package testhelper

import (
	"path/filepath"
	"testing"

	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/adapter/internal/proto_parser"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ReadProtoAsFileDescriptors(t *testing.T, fpath ...string) []*desc.FileDescriptor {
	for i := range fpath {
		fpath[i] = filepath.Join("testdata", fpath[i])
	}
	set, err := proto_parser.ParseFile(fpath, nil)
	require.NoError(t, err)
	assert.Len(t, set, len(fpath))
	return set
}

func FindMessage(t *testing.T, name string, set []*desc.FileDescriptor) *desc.MessageDescriptor {
	for _, f := range set {
		for _, msg := range f.GetMessageTypes() {
			if msg.GetName() == name {
				return msg
			}
		}
	}
	require.Fail(t, "message not found: %s", name)
	return nil
}

func SetupEnv(t *testing.T, fpath, pkgName, svcName string) *entity.Env {
	t.Helper()

	set := helper.ReadProto(t, fpath)

	env := helper.NewEnv(t, set, helper.TestConfig().Env)

	err := env.UsePackage(pkgName)
	require.NoError(t, err)

	if svcName != "" {
		err = env.UseService(svcName)
		require.NoError(t, err)
	}

	return env
}
