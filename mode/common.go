package mode

import (
	"fmt"
	"strings"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/grpc"
	"github.com/ktr0731/evans/grpc/grpcreflection"
	"github.com/ktr0731/evans/idl"
	"github.com/ktr0731/evans/idl/proto"
	"github.com/ktr0731/evans/usecase"
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

func gRPCReflectionPackageFilteredPackages(pkgNames []string) []string {
	pkgs := make([]string, len(pkgNames))
	copy(pkgs, pkgNames)

	n := grpcreflection.ServiceName[:strings.LastIndex(grpcreflection.ServiceName, ".")]
	for i := range pkgs {
		if n == pkgs[i] {
			return append(pkgs[:i], pkgs[i+1:]...)
		}
	}
	return pkgs
}

func setDefault(cfg *config.Config) error {
	// If the spec has only one package, mark it as the default package.
	if cfg.Default.Package == "" {
		pkgs := gRPCReflectionPackageFilteredPackages(usecase.ListPackages())
		if len(pkgs) == 1 {
			cfg.Default.Package = pkgs[0]
		} else {
			hasEmptyPackage := func() bool {
				for _, pkg := range pkgs {
					if pkg == "" {
						return true
					}
				}
				return false
			}()
			if !hasEmptyPackage {
				return nil
			}
		}
	}

	if err := usecase.UsePackage(cfg.Default.Package); err != nil {
		return errors.Wrapf(err, "failed to set '%s' as the default package", cfg.Default.Package)
	}

	// If the spec has only one service, mark it as the default service.
	if cfg.Default.Service == "" {
		svcNames := usecase.ListServicesOld()
		if len(svcNames) != 1 {
			return nil
		}

		cfg.Default.Service = svcNames[0]
		i := strings.LastIndex(cfg.Default.Service, ".")
		if i != -1 {
			cfg.Default.Service = cfg.Default.Service[i+1:]
		}
	}
	if err := usecase.UseService(cfg.Default.Service); err != nil {
		return errors.Wrapf(err, "failed to set '%s' as the default service", cfg.Default.Service)
	}
	return nil
}
