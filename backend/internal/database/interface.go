package database

import (
	"context"
)

type Repo interface {
	InsertDevice(ctx context.Context, name string) (id string, err error)
	AddGeolocation(ctx context.Context, deviceID string, latitude float64, longitude float64) error
}
