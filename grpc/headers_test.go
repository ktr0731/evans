package grpc_test

import (
	"testing"

	"github.com/ktr0731/evans/grpc"
)

func TestHeaders_Add(t *testing.T) {
	cases := map[string]struct {
		k, v   string
		hasErr bool
	}{
		"normal":              {k: "aoi", v: "miyamori"},
		"'/' is invalid char": {k: "aoi/", v: "miyamori", hasErr: true},
	}

	h := grpc.Headers{}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			err := h.Add(c.k, c.v)
			if c.hasErr {
				if err == nil {
					t.Errorf("Add must return an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Add must not return an error, but got '%s'", err)
				}
			}

		})
	}
}