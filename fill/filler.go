// Package fill provides fillers that fills each field with a value.
package fill

import "errors"

var (
	ErrCodecMismatch = errors.New("unsupported codec")
)

// Filler tries to correspond input text to a struct.
type Filler interface {
	// Fill receives a struct v and corresponds input that is have internally to the struct.
	// Fill may return these errors:
	//
	//   - io.EOF: At the end of input.
	//   - ErrCodecMismatch: If v isn't a supported type.
	//
	Fill(v interface{}) error
}
