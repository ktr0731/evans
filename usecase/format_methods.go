package usecase

import (
	"sort"

	"github.com/ktr0731/evans/proto"
	"github.com/pkg/errors"
)

// FormatMethods formats all method names.
func FormatMethods() (string, error) {
	return dm.FormatMethods()
}
func (m *dependencyManager) FormatMethods() (string, error) {
	fqsn := proto.FullyQualifiedServiceName(m.state.selectedPackage, m.state.selectedService)
	v, err := m.methodsToFormatStructs(fqsn)
	if err != nil {
		return "", err
	}
	out, err := m.resourcePresenter.Format(v)
	if err != nil {
		return "", errors.Wrap(err, "failed to format Method names by presenter")
	}
	return out, nil
}

func (m *dependencyManager) methodsToFormatStructs(fqsn string) (v struct {
	Methods []struct {
		Name               string `json:"name" table:"name"`
		FullyQualifiedName string `json:"fully_qualified_name" name:"target" table:"fully-qualified name"`
		RequestType        string `json:"request_type" table:"request type"`
		ResponseType       string `json:"response_type" table:"response type"`
	} `json:"methods" name:"target"`
}, _ error) {
	methods, err := m.listRPCs(fqsn)
	if err != nil {
		return v, errors.Wrapf(err, "failed to list methods associated with '%s'", fqsn)
	}
	for _, m := range methods {
		v.Methods = append(v.Methods, struct {
			Name               string `json:"name" table:"name"`
			FullyQualifiedName string `json:"fully_qualified_name" name:"target" table:"fully-qualified name"`
			RequestType        string `json:"request_type" table:"request type"`
			ResponseType       string `json:"response_type" table:"response type"`
		}{
			Name:               m.Name,
			FullyQualifiedName: m.FullyQualifiedName,
			RequestType:        m.RequestType.FullyQualifiedName,
			ResponseType:       m.ResponseType.FullyQualifiedName,
		})
	}
	sort.Slice(v.Methods, func(i, j int) bool {
		return v.Methods[i].FullyQualifiedName < v.Methods[j].FullyQualifiedName
	})
	return v, nil
}
