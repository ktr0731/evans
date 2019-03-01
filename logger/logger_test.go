package logger_test

import (
	"bytes"
	"testing"

	"github.com/ktr0731/evans/logger"
	"github.com/stretchr/testify/assert"
)

func TestScriptln(t *testing.T) {
	t.Run("logger must write the result of Scriptln to w, but got empty result", func(t *testing.T) {
		defer logger.Reset()
		w := new(bytes.Buffer)
		logger.SetOutput(w)
		logger.Scriptln(func() []interface{} {
			return []interface{}{"aoi", "miyamori"}
		})
		assert.NotEmpty(t, w.String())
	})

	t.Run("logger must not write the result of Scriptln to w because SetOutput is not called", func(t *testing.T) {
		defer logger.Reset()
		w := new(bytes.Buffer)
		logger.Scriptln(func() []interface{} {
			return []interface{}{"erika", "yano"}
		})
		assert.Empty(t, w.String())
	})
}
