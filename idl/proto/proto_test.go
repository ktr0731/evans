package proto_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/idl"
	"github.com/ktr0731/evans/idl/proto"
)

func TestLoadFiles(t *testing.T) {
	_, err := proto.LoadFiles([]string{"testdata"}, []string{"invalid.proto"})
	if err == nil {
		t.Errorf("LoadFiles must return an error, but got nil")
	}

	spec, err := proto.LoadFiles([]string{"testdata"}, []string{"message.proto", "api.proto", "other_package.proto"})
	if err != nil {
		t.Fatalf("LoadFiles must not return an error, but got '%s'", err)
	}

	t.Run("PackageNames", func(t *testing.T) {
		expectedPackageNames := []string{"api", "api2"}
		if diff := cmp.Diff(expectedPackageNames, spec.PackageNames()); diff != "" {
			t.Errorf("PackageNames returned unexpected package names:\n%s", diff)
		}
	})

	t.Run("ServiceNames", func(t *testing.T) {
		_, err = spec.ServiceNames("")
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
		_, err = spec.RPCs("", "")
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
		_, err = spec.RPC("", "", "")
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
		_, err = spec.TypeDescriptor("", "Example")
		if err != idl.ErrPackageUnselected {
			t.Errorf("TypeDescriptor must return ErrPackageUnselected if pkgName is empty, but got '%s'", err)
		}
		_, err := spec.TypeDescriptor("api", "Foo")
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
}
