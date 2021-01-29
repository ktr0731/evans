package usecase

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/grpc"
)

func TestHeader(t *testing.T) {
	defer Clear()
	client, err := grpc.NewClient("", "", false, false, "", "", "", nil)
	if err != nil {
		t.Fatalf("grpc.NewClient must not return an error, but got '%s'", err)
	}
	Inject(Dependencies{GRPCClient: client})

	AddHeader("kumiko", "oumae")

	expected := grpc.Headers{"kumiko": []string{"oumae"}}
	if diff := cmp.Diff(dm.ListHeaders(), expected); diff != "" {
		t.Errorf("unexpected header:\n%s", diff)
	}

	// if a header contains "user-agent" as the header name, it will be ignored.
	AddHeader("User-Agent", "Evans")
	if diff := cmp.Diff(dm.ListHeaders(), expected); diff != "" {
		t.Errorf("unexpected header:\n%s", diff)
	}

	RemoveHeader("kumiko")
	expected = grpc.Headers{}
	if diff := cmp.Diff(dm.ListHeaders(), expected); diff != "" {
		t.Errorf("unexpected header:\n%s", diff)
	}
}
