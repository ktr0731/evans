package fill

import (
	"encoding/json"
	"io"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"
)

// SilentFilter is a Filler implementation that doesn't behave interactive actions.
type SilentFiller struct {
	// dec *json.Decoder
	dec *protojson.UnmarshalOptions
	in  *json.Decoder
}

// NewSilentFiller receives input as io.Reader and returns an instance of SilentFiller.
func NewSilentFiller(in io.Reader) *SilentFiller {
	return &SilentFiller{
		// dec: json.NewDecoder(in),
		dec: &protojson.UnmarshalOptions{
			Resolver: nil, // TODO
		},
		in: json.NewDecoder(in),
	}
}

// Fill fills values of each field from a JSON string. If the JSON string is invalid JSON format or v is a nil pointer,
// Fill returns ErrCodecMismatch.
func (f *SilentFiller) Fill(v *dynamicpb.Message) error {
	var in interface{}
	if err := f.in.Decode(&in); err != nil {
		return err
	}

	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return f.dec.Unmarshal(b, v)
}
