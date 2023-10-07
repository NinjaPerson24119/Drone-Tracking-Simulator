package database

import (
	"time"
)

type Device struct {
	DeviceID string     `json:"device_id" db:"device_id"`
	Name     string     `json:"name" db:"device_name"`
	Created  time.Time  `json:"created" db:"created"`
	Updated  *time.Time `json:"updated" db:"updated"`
	Deleted  *time.Time `json:"deleted" db:"deleted"`
}

type DeviceGeolocation struct {
	DeviceID  string     `json:"device_id" db:"device_id"`
	EventTime time.Time  `json:"event_time" db:"event_time"`
	Latitude  float64    `json:"latitude" db:"latitude"`
	Longitude float64    `json:"longitude" db:"longitude"`
	Created   time.Time  `json:"created" db:"created"`
	Updated   *time.Time `json:"updated" db:"updated"`
	Deleted   *time.Time `json:"deleted" db:"deleted"`
}
