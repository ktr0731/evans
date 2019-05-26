package usecase

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/grpc"
)

func TestHeader(t *testing.T) {
	Clear()

	AddHeader("kumiko", "oumae")

	expected := grpc.Headers{"kumiko": []string{"oumae"}}
	if diff := cmp.Diff(dm.state.headers, expected); diff != "" {
		t.Errorf("unexpected header:\n%s", diff)
	}

	// if a header contains "user-agent" as the header name, it will be ignored.
	AddHeader("User-Agent", "Evans")
	if diff := cmp.Diff(dm.state.headers, expected); diff != "" {
		t.Errorf("unexpected header:\n%s", diff)
	}

	RemoveHeader("kumiko")
	expected = grpc.Headers{}
	if diff := cmp.Diff(dm.state.headers, expected); diff != "" {
		t.Errorf("unexpected header:\n%s", diff)
	}
}
