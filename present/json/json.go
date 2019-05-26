// Package json provides a JSON presenter that formatting.
package json

import (
	gojson "encoding/json"

	"github.com/pkg/errors"
)

// Presenter is a presenter that formats v into JSON string.
type Presenter struct{}

// Format formats v into JSON string. If indent is not empty, Format indents the output.
func (p *Presenter) Format(v interface{}, indent string) (string, error) {
	b, err := gojson.MarshalIndent(v, "", indent)
	if err != nil {
		return "", errors.Wrap(err, "failed to format v into JSON string")
	}
	return string(b), nil
}

func NewPresenter() *Presenter {
	return &Presenter{}
}
