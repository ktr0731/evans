package mode

import (
	"fmt"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/grpc"
	"github.com/ktr0731/evans/grpc/grpcreflection"
	"github.com/ktr0731/evans/idl"
	"github.com/ktr0731/evans/idl/proto"
	"github.com/pkg/errors"
)

func newSpec(cfg *config.Config, grpcClient grpc.Client) (spec idl.Spec, err error) {
	if cfg.Server.Reflection {
		spec, err = proto.LoadByReflection(grpcClient)
	} else {
		spec, err = proto.LoadFiles(cfg.Default.ProtoPath, cfg.Default.ProtoFile)
	}
	if err := errors.Cause(err); err == grpcreflection.ErrTLSHandshakeFailed {
		return nil, errors.New("TLS handshake failed. check whether client or server is misconfigured")
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to instantiate the spec from proto files")
	}
	return spec, nil
}

func newGRPCClient(cfg *config.Config) (grpc.Client, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	if cfg.Request.Web {
		//TODO: remove second arg
		return grpc.NewWebClient(addr, cfg.Server.Reflection, false, "", "", ""), nil
	}
	client, err := grpc.NewClient(
		addr,
		cfg.Server.Name,
		cfg.Server.Reflection,
		cfg.Server.TLS,
		cfg.Request.CACertFile,
		cfg.Request.CertFile,
		cfg.Request.CertKeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to instantiate a gRPC client")
	}
	return client, nil
}
