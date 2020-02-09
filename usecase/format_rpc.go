package usecase

import (
	"github.com/ktr0731/evans/idl"
	"github.com/ktr0731/evans/idl/proto"
	"github.com/pkg/errors"
)

// FormatRPCParams is the modifier for FormatRPC.
type FormatRPCParams struct {
	// FullyQualifiedName means format with a fully-qualified RPC name if it is true.
	FullyQualifiedName bool
}

// FormatRPC formats a RPC.
func FormatRPC(fqrn string, p *FormatRPCParams) (string, error) {
	return dm.FormatRPC(fqrn, p)
}
func (m *dependencyManager) FormatRPC(fqrn string, p *FormatRPCParams) (string, error) {
	if p == nil {
		p = &FormatRPCParams{}
	}

	var v struct {
		Name         string `json:"rpc"`
		RequestType  string `json:"request_type" table:"request type"`
		ResponseType string `json:"response_type" table:"response type"`
	}
	fqsn, _, err := ParseFullyQualifiedMethodName(fqrn)
	if err != nil {
		return "", err
	}
	_, svc := proto.ParseFullyQualifiedServiceName(fqsn)
	rpcs, err := m.ListRPCs(svc)
	if err != nil {
		return "", errors.Wrapf(err, "failed to list RPC associated with '%s'", fqsn)
	}
	for _, r := range rpcs {
		if fqrn != r.FullyQualifiedName {
			continue
		}
		rpcName := r.Name
		if p.FullyQualifiedName {
			fqsn := proto.FullyQualifiedServiceName(m.state.selectedPackage, svc)
			rpcName, err = idl.FullyQualifiedRPCName(fqsn, rpcName)
			if err != nil {
				return "", errors.Wrap(err, "failed to get fully-qualified service name")
			}
		}
		v = struct {
			Name         string `json:"rpc"`
			RequestType  string `json:"request_type" table:"request type"`
			ResponseType string `json:"response_type" table:"response type"`
		}{
			rpcName,
			r.RequestType.Name,
			r.ResponseType.Name,
		}
		out, err := m.resourcePresenter.Format(v)
		if err != nil {
			return "", errors.Wrap(err, "failed to format RPC names by presenter")
		}
		return out, nil
	}
	return "", errors.New("RPC is not found")
}
