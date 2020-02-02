package usecase

import (
	"sort"

	"github.com/ktr0731/evans/idl"
	"github.com/pkg/errors"
)

// FormatRPCsParams is the modifier for FormatRPCs.
type FormatRPCsParams struct {
	// FullyQualifiedName means format with fully-qualified RPC names if it is true.
	FullyQualifiedName bool
}

// FormatRPCs formats all RPC names.
func FormatRPCs(p *FormatRPCsParams) (string, error) {
	return dm.FormatRPCs(p)
}
func (m *dependencyManager) FormatRPCs(p *FormatRPCsParams) (string, error) {
	if p == nil {
		p = &FormatRPCsParams{}
	}

	type rpc struct {
		RPC          string `json:"rpc"`
		RequestType  string `json:"request_type" table:"request type"`
		ResponseType string `json:"response_type" table:"response type"`
	}
	var v struct {
		RPCs []rpc `json:"rpcs"`
	}
	svcName := m.state.selectedService
	rpcs, err := m.ListRPCs(svcName)
	if err != nil {
		return "", errors.Wrapf(err, "failed to list RPCs associated with '%s'", svcName)
	}
	for _, r := range rpcs {
		rpcName := r.Name
		if p.FullyQualifiedName {
			rpcName, err = idl.FullyQualifiedRPCName(m.state.selectedPackage, svcName, rpcName)
			if err != nil {
				return "", errors.Wrap(err, "failed to get fully-qualified service name")
			}
		}
		v.RPCs = append(v.RPCs, rpc{
			rpcName,
			r.RequestType.Name,
			r.ResponseType.Name,
		})
	}
	sort.Slice(v.RPCs, func(i, j int) bool {
		return v.RPCs[i].RPC < v.RPCs[j].RPC
	})
	out, err := m.resourcePresenter.Format(v, "  ")
	if err != nil {
		return "", errors.Wrap(err, "failed to format RPC names by presenter")
	}
	return out, nil
}
