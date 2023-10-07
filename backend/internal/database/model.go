package database

import (
	"time"
)

type Device struct {
	TrackerID string
	Name      float64
	Created   time.Time
	Updated   *time.Time
	Deleted   *time.Time
}

type DeviceGeolocation struct {
	TrackerID string
	EventTime time.Time
	Latitude  float64
	Longitude float64
	Created   time.Time
	Updated   *time.Time
	Deleted   *time.Time
}
