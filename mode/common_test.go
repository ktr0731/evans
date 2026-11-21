package mode

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/proto"
	"github.com/ktr0731/evans/usecase"
)

func TestSetDefaultReflectionVersions(t *testing.T) {
	cases := map[string][]string{
		"v1 only":    {"grpc.reflection.v1.ServerReflection"},
		"alpha only": {"grpc.reflection.v1alpha.ServerReflection"},
		"both":       {"grpc.reflection.v1.ServerReflection", "grpc.reflection.v1alpha.ServerReflection"},
	}
	for name, services := range cases {
		t.Run(name, func(t *testing.T) {
			services := append([]string{"example.Greeter"}, services...)
			usecase.Inject(usecase.Dependencies{DescSource: &proto.DescriptorSourceMock{
				ListServicesFunc: func() ([]string, error) { return services, nil },
			}})
			t.Cleanup(usecase.Clear)

			cfg := &config.Config{Default: &config.Default{}}
			if err := setDefault(cfg); err != nil {
				t.Fatal(err)
			}
			if cfg.Default.Package != "example" || cfg.Default.Service != "Greeter" {
				t.Fatalf("expected example.Greeter as default, got package %q and service %q", cfg.Default.Package, cfg.Default.Service)
			}
		})
	}
}

func TestGRPCReflectionPackageFilteredPackages(t *testing.T) {
	packages := []string{"example", "grpc.reflection.v1", "grpc.reflection.v1alpha", "grpc.reflection.v1.custom"}
	original := append([]string(nil), packages...)
	want := []string{"example", "grpc.reflection.v1.custom"}
	if diff := cmp.Diff(want, gRPCReflectionPackageFilteredPackages(packages)); diff != "" {
		t.Errorf("unexpected packages (-want +got): %s", diff)
	}
	if diff := cmp.Diff(original, packages); diff != "" {
		t.Errorf("input modified (-want +got): %s", diff)
	}
}
