package usecase

import (
	"sort"

	"github.com/pkg/errors"
)

// FormatMessages formats all package names.
func FormatMessages() (string, error) {
	return dm.FormatMessages()
}
func (m *dependencyManager) FormatMessages() (string, error) {
	svcs, err := m.ListServices()
	if err != nil {
		return "", errors.Wrap(err, "failed to list services belong to the package")
	}
	type message struct {
		Message string `json:"message"`
	}
	var v struct {
		Messages []message `json:"messages"`
	}
	encountered := make(map[string]struct{})
	for _, svc := range svcs {
		rpcs, err := m.ListRPCs(svc)
		if err != nil {
			return "", errors.Wrap(err, "failed to list RPCs")
		}
		for _, rpc := range rpcs {
			if _, found := encountered[rpc.RequestType.Name]; !found {
				v.Messages = append(v.Messages, message{rpc.RequestType.Name})
				encountered[rpc.RequestType.Name] = struct{}{}
			}
			if _, found := encountered[rpc.ResponseType.Name]; !found {
				v.Messages = append(v.Messages, message{rpc.ResponseType.Name})
				encountered[rpc.ResponseType.Name] = struct{}{}
			}
		}
	}
	sort.Slice(v.Messages, func(i, j int) bool {
		return v.Messages[i].Message < v.Messages[j].Message
	})
	out, err := m.resourcePresenter.Format(v, "  ")
	if err != nil {
		return "", errors.Wrap(err, "failed to format message names by presenter")
	}
	return out, nil
}
