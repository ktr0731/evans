package port

import (
	"context"
)

type GRPCPort interface {
	Invoke(ctx context.Context, fqrn string, req, res interface{}) error
}
