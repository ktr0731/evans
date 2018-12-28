package testhelper

import (
	"path/filepath"
	"testing"

	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/adapter/internal/protoparser"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/entity/env"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/require"
)

func ReadProtoAsFileDescriptors(t *testing.T, fpath ...string) []*desc.FileDescriptor {
	for i := range fpath {
		fpath[i] = filepath.Join("testdata", fpath[i])
	}
	set, err := protoparser.ParseFile(fpath, nil)
	require.NoError(t, err)
	require.Len(t, set, len(fpath))
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

func SetupEnv(t *testing.T, fpath, pkgName, svcName string) *env.Env {
	t.Helper()

	set := helper.ReadProto(t, fpath)

	cfg, err := config.Get(nil)
	require.NoError(t, err, "failed to get a config")

	headers := make([]entity.Header, 0, len(cfg.Request.Header))
	for k, v := range cfg.Request.Header {
		require.Len(t, v, 1, "currently, header length is always 1")
		headers = append(headers, entity.Header{Key: k, Val: v[0]})
	}
	env := env.New(set, headers)

	err = env.UsePackage(pkgName)
	require.NoError(t, err)

	if svcName != "" {
		err = env.UseService(svcName)
		require.NoError(t, err)
	}

	return env
}
