package database

import (
	"context"
)

type Repo interface {
	InsertDevice(ctx context.Context, device *Device) (string, error)
	InsertGeolocation(ctx context.Context, geolocation *DeviceGeolocation) error
	GetLatestGeolocations(ctx context.Context, page int, pageSize int) ([]*DeviceGeolocation, error)
}
