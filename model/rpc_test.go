package model

import (
	"fmt"
	"testing"

	"github.com/jhump/protoreflect/desc"
	"github.com/stretchr/testify/assert"
)

func TestNewRPCs(t *testing.T) {
	const pkgName = "steinsgate"
	fd := fileDesc(t, []string{"testdata/test.proto"}, []string{})

	msg := func(msgName string) *desc.MessageDescriptor {
		return getMessage(t, fd, pkgName, msgName).Desc
	}

	tests := map[string]struct {
		svcName string
		expect  RPCs
		err     error
	}{
		"normal": {
			svcName: "Action",
			expect: RPCs{
				{Name: "Timeleap", RequestType: msg("TimeleapReq"), ResponseType: msg("TimeleapRes")},
			},
		},
	}

	for title, test := range tests {
		t.Run(title, func(t *testing.T) {
			svc := getService(t, fd, pkgName, test.svcName)

			actual := NewRPCs(svc)
			assert.Exactly(t, test.expect, actual)
		})
	}
}

func TestRPCs_String(t *testing.T) {
	const pkgName = "steinsgate"
	fd := fileDesc(t, []string{"testdata/test.proto"}, []string{})

	tests := map[string]struct {
		svcName string
		expect  string
	}{
		"normal": {svcName: "Action"},
	}

	for title, test := range tests {
		t.Run(title, func(t *testing.T) {
			svc := getService(t, fd, pkgName, test.svcName)

			rpcs := NewRPCs(svc)
			actual := rpcs.String()

			fmt.Println(actual)
		})
	}
}
