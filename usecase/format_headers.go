package usecase

import (
	"sort"

	"github.com/pkg/errors"
)

// FormatHeaders formats all package names.
func FormatHeaders() (string, error) {
	return dm.FormatHeaders()
}
func (m *dependencyManager) FormatHeaders() (string, error) {
	type header struct {
		Key string `json:"key"`
		Val string `json:"val"`
	}
	var s struct {
		Headers []header `json:"headers"`
	}
	headers := m.ListHeaders()
	for k, v := range headers {
		for _, vv := range v {
			s.Headers = append(s.Headers, header{k, vv})
		}
	}
	sort.Slice(s.Headers, func(i, j int) bool {
		return s.Headers[i].Key < s.Headers[j].Key
	})
	out, err := m.resourcePresenter.Format(s)
	if err != nil {
		return "", errors.Wrap(err, "failed to format header names by presenter")
	}
	return out, nil
}
