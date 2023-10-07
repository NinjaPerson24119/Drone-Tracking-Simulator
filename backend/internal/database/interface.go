package database

import (
	"context"

	"github.com/NinjaPerson24119/MapProject/backend/internal/filters"
)

type Repo interface {
	InsertDevice(ctx context.Context, device *Device) (string, error)
	ListDevices(ctx context.Context, paging filters.PageOptions) ([]*Device, error)
	InsertGeolocation(ctx context.Context, geolocation *DeviceGeolocation) error
	GetLatestGeolocations(ctx context.Context, paging filters.PageOptions) ([]*DeviceGeolocation, error)
}
