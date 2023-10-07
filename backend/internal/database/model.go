package database

import (
	"time"
)

type Device struct {
	DeviceID string     `json:"device_id"`
	Name     float64    `json:"name"`
	Created  time.Time  `json:"created"`
	Updated  *time.Time `json:"updated"`
	Deleted  *time.Time `json:"deleted"`
}

type DeviceGeolocation struct {
	DeviceID  string     `json:"device_id"`
	EventTime time.Time  `json:"event_time"`
	Latitude  float64    `json:"latitude"`
	Longitude float64    `json:"longitude"`
	Created   time.Time  `json:"created"`
	Updated   *time.Time `json:"updated"`
	Deleted   *time.Time `json:"deleted"`
}
