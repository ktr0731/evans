package logger_test

import (
	"bytes"
	"testing"

	"github.com/ktr0731/evans/logger"
)

func TestScriptln(t *testing.T) {
	t.Run("logger must write the result of Scriptln/Scriptf to w", func(t *testing.T) {
		defer logger.Reset()
		w := new(bytes.Buffer)
		logger.SetOutput(w)
		logger.Scriptln(func() []interface{} {
			return []interface{}{"aoi", "miyamori"}
		})
		if w.Len() == 0 {
			t.Errorf("Scriptln must write the result to w, but empty")
		}
		if expected := "evans: aoi miyamori\n"; w.String() != expected {
			t.Errorf("expected = '%s', but got '%s'", expected, w.String())
		}

		w.Truncate(0)

		logger.Scriptf("%s-%s", func() []interface{} {
			return []interface{}{"aoi", "miyamori"}
		})
		if w.Len() == 0 {
			t.Errorf("Scriptf must write the result to w, but empty")
		}
		if expected := "evans: aoi-miyamori\n"; w.String() != expected {
			t.Errorf("expected = '%s', but got '%s'", expected, w.String())
		}
	})

	t.Run("logger must not write the result of Scriptln to w because SetOutput is not called", func(t *testing.T) {
		defer logger.Reset()
		w := new(bytes.Buffer)
		logger.Scriptln(func() []interface{} {
			return []interface{}{"erika", "yano"}
		})
		if w.Len() != 0 {
			t.Errorf("Scriptln must not write the result to w")
		}
	})
}
