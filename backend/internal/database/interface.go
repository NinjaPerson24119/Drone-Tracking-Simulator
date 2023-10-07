package database

import (
	"context"
	"time"
)

type Repo interface {
	InsertDevice(ctx context.Context, name string) (string, error)
	InsertGeolocation(ctx context.Context, deviceID string, eventTime time.Time, latitude float64, longitude float64)
	GetLatestGeolocations(ctx context.Context, deviceID string) (string, time.Time, float64, float64, error)
}
