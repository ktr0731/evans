package mode

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func Test_tidyUpHistory(t *testing.T) {
	cases := map[string]struct {
		history     []string
		historySize int
		expected    []string
	}{
		"empty": {
			history:     nil,
			historySize: 100,
			expected:    []string{},
		},
		"simple": {
			history:     []string{"foo", "bar"},
			historySize: 100,
			expected:    []string{"foo", "bar"},
		},
		"remove duplicated items": {
			history:     []string{"foo", "bar", "foo", "baz"},
			historySize: 100,
			expected:    []string{"bar", "foo", "baz"},
		},
		"over history size": {
			history:     []string{"foo", "bar", "baz"},
			historySize: 2,
			expected:    []string{"bar", "baz"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := tidyUpHistory(c.history, c.historySize)
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("-want, +got\n%s", diff)
			}
		})
	}
}

func Test_curlLikeResponsePresenter(t *testing.T) {
	ok := codes.OK
	cases := map[string]struct {
		res      *usecase.GRPCResponse
		expected string
		hasErr   bool
	}{
		"only body": {
			res: &usecase.GRPCResponse{
				Message: struct {
					Name string `json:"name"`
				}{Name: "oumae"},
			},
			expected: `{
  "name": "oumae"
}`,
		},
		"status and body": {
			res: &usecase.GRPCResponse{
				Status: &ok,
				Message: struct {
					Name string `json:"name"`
				}{Name: "oumae"},
			},
			expected: `0 OK

{
  "name": "oumae"
}`,
		},
		"status, header and body": {
			res: &usecase.GRPCResponse{
				Status: &ok,
				HeaderMetadata: &metadata.MD{
					"oumae":   []string{"kumiko", "mamiko"},
					"kousaka": []string{"reina"},
				},
				Message: struct {
					Name string `json:"name"`
				}{Name: "oumae"},
			},
			expected: `0 OK
kousaka: reina
oumae: kumiko
oumae: mamiko

{
  "name": "oumae"
}`,
		},
		"status, header, body and trailer": {
			res: &usecase.GRPCResponse{
				Status: &ok,
				HeaderMetadata: &metadata.MD{
					"oumae":   []string{"kumiko", "mamiko"},
					"kousaka": []string{"reina"},
				},
				Message: struct {
					Name string `json:"name"`
				}{Name: "oumae"},
				TrailerMetadata: &metadata.MD{
					"kato":      []string{"hazuki"},
					"kawashima": []string{"sapphire"},
				},
			},
			expected: `0 OK
kousaka: reina
oumae: kumiko
oumae: mamiko

{
  "name": "oumae"
}

kato: hazuki
kawashima: sapphire`,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p := newCurlLikeResponsePresenter()
			actual, err := p.Format(c.res)
			if c.hasErr {
				if err == nil {
					t.Errorf("should return an error, but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("should not return an error, but got '%s'", err)
			}
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("-want, +got\n%s", diff)
			}
		})
	}
}
