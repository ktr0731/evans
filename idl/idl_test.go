package idl_test

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/idl"
	"github.com/ktr0731/evans/idl/proto"
)

func TestSpec(t *testing.T) {
	cases := map[string]struct {
		newNormalSpec       func(*testing.T) idl.Spec
		newEmptyPackageSpec func(*testing.T) idl.Spec
	}{
		"proto": {
			newNormalSpec: func(t *testing.T) idl.Spec {
				fnames := []string{"message.proto", "api.proto", "other_package.proto"}
				spec, err := proto.LoadFiles([]string{filepath.Join("proto", "testdata")}, fnames)
				if err != nil {
					t.Fatalf("LoadFiles must not return an error, but got '%s'", err)
				}
				return spec
			},
			newEmptyPackageSpec: func(t *testing.T) idl.Spec {
				fnames := []string{"empty_package.proto"}
				spec, err := proto.LoadFiles([]string{filepath.Join("proto", "testdata")}, fnames)
				if err != nil {
					t.Fatalf("LoadFiles must not return an error, but got '%s'", err)
				}
				return spec
			},
		},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			t.Run("normal", func(t *testing.T) {
				spec := c.newNormalSpec(t)

				t.Run("ServiceNames", func(t *testing.T) {
					actualServiceNames := spec.ServiceNames()
					expectedServiceNames := []string{"api.Example"}
					if diff := cmp.Diff(expectedServiceNames, actualServiceNames); diff != "" {
						t.Errorf("ServiceNames returned unexpected service names:\n%s", diff)
					}
				})

				t.Run("RPCs", func(t *testing.T) {
					_, err := spec.RPCs("")
					if err != idl.ErrServiceUnselected {
						t.Errorf("RPCs must return ErrServiceUnselected if svcName is empty, but got '%s'", err)
					}
					_, err = spec.RPCs("Foo")
					if err != idl.ErrUnknownServiceName {
						t.Errorf("RPCs must return ErrUnknownServiceName, but got '%s'", err)
					}

					rpcs, err := spec.RPCs("api.Example")
					if err != nil {
						t.Fatalf("api.Example service must have an RPC, but couldn't get it: '%s'", err)
					}
					actualRPCNames := make([]string, len(rpcs))
					for i, rpc := range rpcs {
						actualRPCNames[i] = rpc.Name
					}
					expectedRPCNames := []string{"RPC"}
					if diff := cmp.Diff(expectedRPCNames, actualRPCNames); diff != "" {
						t.Errorf("RPCs returned unexpected RPC names:\n%s", diff)
					}
				})

				t.Run("RPC", func(t *testing.T) {
					_, err := spec.RPC("", "")
					if err != idl.ErrServiceUnselected {
						t.Errorf("RPC must return ErrServiceUnselected if svcName is empty, but got '%s'", err)
					}
					_, err = spec.RPC("Foo", "")
					if err != idl.ErrUnknownServiceName {
						t.Errorf("RPC must return ErrUnknownServiceName, but got '%s'", err)
					}
					_, err = spec.RPC("api.Example", "")
					if err != idl.ErrUnknownRPCName {
						t.Errorf("RPC must return ErrUnknownRPCName if rpcName is empty, but got '%s'", err)
					}

					actualRPC, err := spec.RPC("api.Example", "RPC")
					if err != nil {
						t.Fatalf("Example service must have an RPC named 'RPC', but couldn't get it: '%s'", err)
					}

					const expectedFQRN = "api.Example.RPC"
					if actualFQRN := actualRPC.FullyQualifiedName; actualFQRN != expectedFQRN {
						t.Errorf("expected FullyQualifiedName is '%s', but got '%s'", expectedFQRN, actualFQRN)
					}
				})

				t.Run("TypeDescriptor", func(t *testing.T) {
					_, err := spec.TypeDescriptor("Foo")
					if err == nil {
						t.Fatalf("TypeDescriptor must return an error because api.Foo is an undefined type, but got nil")
					}
					actual, err := spec.TypeDescriptor("api.Request")
					if err != nil {
						t.Fatalf("TypeDescriptor must return the descriptor of api.Request, but got an error: '%s'", err)
					}

					if actual == nil {
						t.Errorf("actual must not be nil")
					}
				})
			})

			t.Run("empty package", func(t *testing.T) {
				spec := c.newEmptyPackageSpec(t)

				t.Run("ServiceNames", func(t *testing.T) {
					actualServiceNames := spec.ServiceNames()
					expectedServiceNames := []string{"Example"}
					if diff := cmp.Diff(expectedServiceNames, actualServiceNames); diff != "" {
						t.Errorf("ServiceNames returned unexpected service names:\n%s", diff)
					}
				})

				t.Run("RPCs", func(t *testing.T) {
					_, err := spec.RPCs("")
					if err != idl.ErrServiceUnselected {
						t.Errorf("RPCs must return ErrServiceUnselected if svcName is empty, but got '%s'", err)
					}
					_, err = spec.RPCs("Foo")
					if err != idl.ErrUnknownServiceName {
						t.Errorf("RPCs must return ErrUnknownServiceName, but got '%s'", err)
					}
					rpcs, err := spec.RPCs("Example")
					if err != nil {
						t.Fatalf("RPCs must not return an error if pkgName is empty, but got '%s'", err)
					}

					actualRPCNames := make([]string, len(rpcs))
					for i, rpc := range rpcs {
						actualRPCNames[i] = rpc.Name
					}
					expectedRPCNames := []string{"RPC"}
					if diff := cmp.Diff(expectedRPCNames, actualRPCNames); diff != "" {
						t.Errorf("RPCs returned unexpected RPC names:\n%s", diff)
					}
				})

				t.Run("RPC", func(t *testing.T) {
					_, err := spec.RPC("Example", "")
					if err != idl.ErrUnknownRPCName {
						t.Errorf("RPC must return ErrUnknownRPCName if rpcName is empty, but got '%s'", err)
					}

					actualRPC, err := spec.RPC("Example", "RPC")
					if err != nil {
						t.Fatalf("Example service must have an RPC named 'RPC', but couldn't get it: '%s'", err)
					}

					const expectedFQRN = "Example.RPC"
					if actualFQRN := actualRPC.FullyQualifiedName; actualFQRN != expectedFQRN {
						t.Errorf("expected FullyQualifiedName is '%s', but got '%s'", expectedFQRN, actualFQRN)
					}
				})

				t.Run("TypeDescriptor", func(t *testing.T) {
					actual, err := spec.TypeDescriptor("Request")
					if err != nil {
						t.Fatalf("TypeDescriptor must return the descriptor of api.Request, but got an error: '%s'", err)
					}

					if actual == nil {
						t.Errorf("actual must not be nil")
					}
				})
			})
		})
	}
}
