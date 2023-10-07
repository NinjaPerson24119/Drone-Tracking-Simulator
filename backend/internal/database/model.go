package database

import (
	"time"
)

type Device struct {
	DeviceID string
	Name     float64
	Created  time.Time
	Updated  *time.Time
	Deleted  *time.Time
}

type DeviceGeolocation struct {
	DeviceID  string
	EventTime time.Time
	Latitude  float64
	Longitude float64
	Created   time.Time
	Updated   *time.Time
	Deleted   *time.Time
}
