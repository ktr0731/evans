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

// InteractiveFillerOpts
// If DigManually is true, Fill asks whether to dig down if it encountered to a message field.
// If BytesFromFile is true, Fill will read the contents of the file from the provided relative path
type InteractiveFillerOpts struct {
	DigManually, BytesFromFile bool
}

// Filler tries to correspond input text to a struct interactively.
type InteractiveFiller interface {
	// Fill receives a struct v and corresponds input that is have internally to the struct.
	// Fill may return these errors:
	//
	//   - io.EOF: At the end of input.
	//   - ErrCodecMismatch: If v isn't a supported type.
	//
	Fill(v interface{}, opts InteractiveFillerOpts) error
}
