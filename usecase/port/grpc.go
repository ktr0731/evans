package port

import (
	"context"
)

type GRPCPort interface {
	Invoke(ctx context.Context, req, res interface{})
}
