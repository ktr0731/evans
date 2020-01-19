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

				t.Run("PackageNames", func(t *testing.T) {
					expectedPackageNames := []string{"api", "api2"}
					if diff := cmp.Diff(expectedPackageNames, spec.PackageNames()); diff != "" {
						t.Errorf("PackageNames returned unexpected package names:\n%s", diff)
					}
				})

				t.Run("ServiceNames", func(t *testing.T) {
					_, err := spec.ServiceNames("")
					if err != idl.ErrPackageUnselected {
						t.Errorf("ServiceNames must return ErrPackageUnselected if pkgName is empty, but got '%s'", err)
					}
					_, err = spec.ServiceNames("foo")
					if err != idl.ErrUnknownPackageName {
						t.Errorf("ServiceNames must return ErrUnknownPackageName, but got '%s'", err)
					}

					actualServiceNames, err := spec.ServiceNames("api")
					if err != nil {
						t.Fatalf("api package must have a service, but couldn't get it: '%s'", err)
					}
					expectedServiceNames := []string{"Example"}
					if diff := cmp.Diff(expectedServiceNames, actualServiceNames); diff != "" {
						t.Errorf("ServiceNames returned unexpected service names:\n%s", diff)
					}
				})

				t.Run("RPCs", func(t *testing.T) {
					_, err := spec.RPCs("", "")
					if err != idl.ErrPackageUnselected {
						t.Errorf("RPCs must return ErrPackageUnselected if pkgName is empty, but got '%s'", err)
					}
					_, err = spec.RPCs("api", "")
					if err != idl.ErrServiceUnselected {
						t.Errorf("RPCs must return ErrServiceUnselected if svcName is empty, but got '%s'", err)
					}
					_, err = spec.RPCs("foo", "")
					if err != idl.ErrUnknownPackageName {
						t.Errorf("RPCs must return ErrUnknownPackageName, but got '%s'", err)
					}
					_, err = spec.RPCs("api", "Foo")
					if err != idl.ErrUnknownServiceName {
						t.Errorf("RPCs must return ErrUnknownServiceName, but got '%s'", err)
					}

					rpcs, err := spec.RPCs("api", "Example")
					if err != nil {
						t.Fatalf("Example service of api package must have an RPC, but couldn't get it: '%s'", err)
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
					_, err := spec.RPC("", "", "")
					if err != idl.ErrPackageUnselected {
						t.Errorf("RPC must return ErrPackageUnselected if pkgName is empty, but got '%s'", err)
					}
					_, err = spec.RPC("api", "", "")
					if err != idl.ErrServiceUnselected {
						t.Errorf("RPC must return ErrServiceUnselected if svcName is empty, but got '%s'", err)
					}
					_, err = spec.RPC("foo", "", "")
					if err != idl.ErrUnknownPackageName {
						t.Errorf("RPC must return ErrUnknownPackageName, but got '%s'", err)
					}
					_, err = spec.RPC("api", "Foo", "")
					if err != idl.ErrUnknownServiceName {
						t.Errorf("RPC must return ErrUnknownServiceName, but got '%s'", err)
					}
					_, err = spec.RPC("api", "Example", "")
					if err != idl.ErrUnknownRPCName {
						t.Errorf("RPC must return ErrUnknownRPCName if rpcName is empty, but got '%s'", err)
					}

					actualRPC, err := spec.RPC("api", "Example", "RPC")
					if err != nil {
						t.Fatalf("Example service of api package must have an RPC named 'RPC', but couldn't get it: '%s'", err)
					}

					const expectedFQRN = "api.Example.RPC"
					if actualFQRN := actualRPC.FullyQualifiedName; actualFQRN != expectedFQRN {
						t.Errorf("expected FullyQualifiedName is '%s', but got '%s'", expectedFQRN, actualFQRN)
					}
				})

				t.Run("TypeDescriptor", func(t *testing.T) {
					_, err := spec.TypeDescriptor("", "Request")
					if err != idl.ErrPackageUnselected {
						t.Errorf("TypeDescriptor must return ErrPackageUnselected if pkgName is empty, but got '%s'", err)
					}
					_, err = spec.TypeDescriptor("api", "Foo")
					if err == nil {
						t.Fatalf("TypeDescriptor must return an error because api.Foo is an undefined type, but got nil")
					}
					actual, err := spec.TypeDescriptor("api", "Request")
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

				t.Run("PackageNames", func(t *testing.T) {
					expectedPackageNames := []string{""}
					if diff := cmp.Diff(expectedPackageNames, spec.PackageNames()); diff != "" {
						t.Errorf("PackageNames returned unexpected package names:\n%s", diff)
					}
				})

				t.Run("ServiceNames", func(t *testing.T) {
					actualServiceNames, err := spec.ServiceNames("")
					if err != nil {
						t.Fatalf("ServiceNames must not return error because package is unspecified, but got '%s'", err)
					}
					expectedServiceNames := []string{"Example"}
					if diff := cmp.Diff(expectedServiceNames, actualServiceNames); diff != "" {
						t.Errorf("ServiceNames returned unexpected service names:\n%s", diff)
					}
				})

				t.Run("RPCs", func(t *testing.T) {
					_, err := spec.RPCs("", "")
					if err != idl.ErrServiceUnselected {
						t.Errorf("RPCs must return ErrServiceUnselected if svcName is empty, but got '%s'", err)
					}
					_, err = spec.RPCs("", "Foo")
					if err != idl.ErrUnknownServiceName {
						t.Errorf("RPCs must return ErrUnknownServiceName, but got '%s'", err)
					}
					rpcs, err := spec.RPCs("", "Example")
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
					_, err := spec.RPC("", "Example", "")
					if err != idl.ErrUnknownRPCName {
						t.Errorf("RPC must return ErrUnknownRPCName if rpcName is empty, but got '%s'", err)
					}

					actualRPC, err := spec.RPC("", "Example", "RPC")
					if err != nil {
						t.Fatalf("Example service of the package must have an RPC named 'RPC', but couldn't get it: '%s'", err)
					}

					const expectedFQRN = "Example.RPC"
					if actualFQRN := actualRPC.FullyQualifiedName; actualFQRN != expectedFQRN {
						t.Errorf("expected FullyQualifiedName is '%s', but got '%s'", expectedFQRN, actualFQRN)
					}
				})

				t.Run("TypeDescriptor", func(t *testing.T) {
					actual, err := spec.TypeDescriptor("", "Request")
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

func TestFullyQualifiedServiceName(t *testing.T) {
	cases := map[string]struct {
		pkg, svc    string
		expected    string
		expectedErr error
	}{
		"normal":             {pkg: "foo", svc: "Bar", expected: "foo.Bar"},
		"only service":       {svc: "Bar", expected: "Bar"},
		"service unselected": {pkg: "foo", expectedErr: idl.ErrServiceUnselected},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			fqsn, err := idl.FullyQualifiedServiceName(c.pkg, c.svc)
			if c.expectedErr != nil {
				if err != c.expectedErr {
					t.Errorf("expected error '%s', but got '%s'", c.expectedErr, err)
				}
				return
			}
			if fqsn != c.expected {
				t.Errorf("expected %s, but got %s", c.expected, fqsn)
			}
		})
	}
}
