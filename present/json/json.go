// Package json provides a JSON presenter that formatting.
package json

import (
	gojson "encoding/json"

	"github.com/pkg/errors"
)

// Presenter is a presenter that formats v into JSON string.
type Presenter struct {
	indent string
}

// Format formats v into JSON string.
func (p *Presenter) Format(v interface{}) (string, error) {
	b, err := gojson.MarshalIndent(v, "", p.indent)
	if err != nil {
		return "", errors.Wrap(err, "failed to format v into JSON string")
	}
	return string(b), nil
}

// NewPresenter instantiates a JSON presenter.
// If indent is not empty, Format indents the output.
func NewPresenter(indent string) *Presenter {
	return &Presenter{indent: indent}
}
