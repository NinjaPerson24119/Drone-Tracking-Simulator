package tracker

import (
	"context"
)

type Service interface {
	InsertDevice(ctx context.Context, name string) (id string, error)
	AddGeolocation(ctx context.Context, deviceID string, latitude float64, longitude float64) error
}
