package simulator

import (
	"context"
)

type Simulator interface {
	Run(ctx context.Context) error
}
