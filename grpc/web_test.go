package grpc_test

import (
	"context"
	"testing"

	"github.com/ktr0731/evans/grpc"
)

func TestWebClient(t *testing.T) {
	client := grpc.NewWebClient("", false, false, "", "", "")
	t.Run("Invoke returns an error if FQRN is invalid", func(t *testing.T) {
		_, _, err := client.Invoke(context.Background(), "invalid-fqrn", nil, nil)
		if err == nil {
			t.Errorf("expected an error, but got nil")
		}
	})
	t.Run("NewClientStream returns an error if FQRN is invalid", func(t *testing.T) {
		_, err := client.NewClientStream(context.Background(), nil, "invalid-fqrn")
		if err == nil {
			t.Errorf("expected an error, but got nil")
		}
	})
	t.Run("NewServerStream returns an error if FQRN is invalid", func(t *testing.T) {
		_, err := client.NewServerStream(context.Background(), nil, "invalid-fqrn")
		if err == nil {
			t.Errorf("expected an error, but got nil")
		}
	})
	t.Run("NewBidiStream returns an error if FQRN is invalid", func(t *testing.T) {
		_, err := client.NewBidiStream(context.Background(), nil, "invalid-fqrn")
		if err == nil {
			t.Errorf("expected an error, but got nil")
		}
	})
}
