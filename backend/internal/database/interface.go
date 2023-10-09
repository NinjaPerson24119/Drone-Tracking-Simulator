package database

import (
	"context"

	"github.com/NinjaPerson24119/MapProject/backend/internal/filters"
)

type Repo interface {
	Close()
	InsertDevice(ctx context.Context, device *Device) (string, error)
	ListDevices(ctx context.Context, paging filters.PageOptions) ([]*Device, error)
	InsertGeolocation(ctx context.Context, geolocation *DeviceGeolocation) error
	ListLatestGeolocations(ctx context.Context, paging filters.PageOptions) ([]*DeviceGeolocation, error)
	GetMultiLatestGeolocations(ctx context.Context, deviceIDs []string) ([]*DeviceGeolocation, error)
	ListenToGeolocationInserted(ctx context.Context, handler func(string) error) error
}
