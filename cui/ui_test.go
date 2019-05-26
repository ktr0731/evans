package cui

import (
	"testing"
)

func TestDefaultUI(t *testing.T) {
	ui := DefaultUI()
	if *ui != defaultUI {
		t.Errorf("field value mismatch: expected = %s, actual = %s", *ui, defaultUI)
	}
	if ui == &defaultUI {
		t.Error("the returned value must not have the reference that is the same as the defaultUI")
	}
}
