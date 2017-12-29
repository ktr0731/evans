package entity

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	const pkgName = "steinsgate"
	fd := fileDesc(t, []string{"testdata/test.proto"}, []string{})
	expect := getMessage(t, fd, pkgName, "TimeleapReq")
	actual := NewMessage(expect.Desc)
	assert.Exactly(t, expect, actual)
}

func TestMessage_String(t *testing.T) {
	const pkgName = "steinsgate"
	fd := fileDesc(t, []string{"testdata/test.proto"}, []string{})
	msg := NewMessage(getMessage(t, fd, pkgName, "TimeleapReq").Desc)
	fmt.Println(msg.String())
}

func TestMessages_String(t *testing.T) {
	const pkgName = "steinsgate"
	fd := fileDesc(t, []string{"testdata/test.proto"}, []string{})
	msg := &Messages{NewMessage(getMessage(t, fd, pkgName, "TimeleapReq").Desc)}
	fmt.Println(msg.String())
}
