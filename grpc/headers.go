package grpc

import (
	"unicode"

	"github.com/pkg/errors"
)

// Headers represents gRPC headers.
// A key corresponds to one or more values.
type Headers map[string][]string

// Add appends a value v to a key k.
// k must be consisted of other than '-', '_' and '.'.
func (h Headers) Add(k, v string) error {
	// If k is already in h, k is valid key name.
	if _, ok := h[k]; !ok {
		for _, r := range k {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' && r != '.' {
				return errors.Errorf("invalid char '%c' in key", r)
			}
		}
	}
	h[k] = append(h[k], v)
	return nil
}

// Remove deletes values corresponds to a key k.
func (h Headers) Remove(k string) {
	delete(h, k)
}
