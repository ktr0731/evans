package entity

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	const pkgName = "steinsgate"
	fd := fileDesc(t, []string{"testdata/test.proto"}, []string{})

	t.Run("normal", func(t *testing.T) {
		svc := getService(t, fd, pkgName, "Action")
		actual := NewService(svc)

		rpcNum := len(svc.GetMethods())
		rpcs := make(RPCs, rpcNum)
		for i, rpc := range svc.GetMethods() {
			rpcs[i] = &RPC{
				Name:         rpc.GetName(),
				RequestType:  rpc.GetInputType(),
				ResponseType: rpc.GetOutputType(),
			}
		}

		expected := &Service{
			Name: "Action",
			RPCs: rpcs,
		}

		assert.Exactly(t, expected, actual)
	})
}

func TestServices_String(t *testing.T) {
	const pkgName = "steinsgate"
	fd := fileDesc(t, []string{"testdata/test.proto"}, []string{})
	svcDesc := getService(t, fd, pkgName, "Action")
	svc := &Services{NewService(svcDesc)}
	fmt.Println(svc)
}
