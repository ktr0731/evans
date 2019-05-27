package fill

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

// SilentFilter is a Filler implementation that doesn't behave interactive actions.
type SilentFiller struct {
	dec *json.Decoder
}

// NewSilentFiller receives input as io.Reader and returns an instance of SilentFiller.
func NewSilentFiller(in io.Reader) *SilentFiller {
	return &SilentFiller{
		dec: json.NewDecoder(in),
	}
}

// Fill fills values of each field from a JSON string. If the JSON string is invalid JSON format or v is a nil pointer,
// Fill returns ErrCodecMismatch.
func (f *SilentFiller) Fill(v interface{}) error {
	err := f.dec.Decode(v)
	if err != nil {
		if err == io.EOF {
			return io.EOF
		}

		switch err.(type) {
		case *json.InvalidUnmarshalError, *json.SyntaxError:
			return ErrCodecMismatch
		default:
			return errors.Wrap(err, "failed to read input as JSON")
		}
	}
	return nil
}
