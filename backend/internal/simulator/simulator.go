package simulator

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/NinjaPerson24119/MapProject/backend/internal/database"
	"github.com/NinjaPerson24119/MapProject/backend/internal/filters"
)

type SimulatedDevice struct {
	device            *database.Device
	geolocation       *database.DeviceGeolocation
	stepDisplacementX float64
	stepDisplacementY float64
	// if this is >0, we won't switch directions until N steps have passed
	switchDirectionCooldownSteps int
	lastUpdate                   time.Time
}

type SimulatorImpl struct {
	repo database.Repo

	simulatedDevices []*SimulatedDevice

	noDevices       int
	centerLatitude  float64
	centerLongitude float64
	radius          float64
	frequency       float64
	sleepTime       time.Duration
	movementPerSec  float64

	maxInsertRetries int
	insertRetryTime  time.Duration
}

func New(repo database.Repo, noDevices int, centerLatitude float64, centerLongitude float64, radius float64, frequency float64, movementPerSec float64) *SimulatorImpl {
	return &SimulatorImpl{
		repo:             repo,
		noDevices:        noDevices,
		centerLatitude:   centerLatitude,
		centerLongitude:  centerLongitude,
		radius:           radius,
		frequency:        frequency,
		sleepTime:        time.Duration(1.0 / frequency * float64(time.Second)),
		movementPerSec:   movementPerSec,
		maxInsertRetries: 5,
		insertRetryTime:  time.Duration(2 * time.Millisecond),
	}
}

func (s *SimulatorImpl) Run(ctx context.Context) error {
	err := s.setupDevices(ctx)
	if err != nil {
		return err
	}

	// simulate movement
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			before := time.Now()
			err := s.stepDevices(ctx)
			if err != nil {
				break
			}
			after := time.Now()
			delta := after.Sub(before)
			if delta > s.sleepTime {
				fmt.Printf("stepDevices exhausted allocated step time: %v > %v\n", delta, s.sleepTime)
			}
			time.Sleep(s.sleepTime - delta)
		}
	}
}

func (s *SimulatorImpl) setupDevices(ctx context.Context) error {
	// fetch devices
	devices, err := s.repo.ListDevices(ctx, filters.PageOptions{
		Page:     1,
		PageSize: s.noDevices,
	})
	if err != nil {
		return err
	}

	// ensure minimum number of devices
	if len(devices) < s.noDevices {
		for i := len(devices); i < s.noDevices; i++ {
			device := &database.Device{
				Name: fmt.Sprintf("ReallyBigTruck-%d", i),
			}
			id, err := s.repo.InsertDevice(ctx, device)
			if err != nil {
				return err
			}
			device.DeviceID = id
			devices = append(devices, device)
		}
	}

	// pick random starting locations and directions
	for _, device := range devices {
		deviceGeolocation := &database.DeviceGeolocation{
			DeviceID:  device.DeviceID,
			EventTime: time.Now(),
			Latitude:  s.centerLatitude + s.radius/2*(rand.Float64()-0.5)*2,
			Longitude: s.centerLongitude + s.radius/2*(rand.Float64()-0.5)*2,
		}

		directionRadians := 2 * math.Pi * rand.Float64()
		s.simulatedDevices = append(s.simulatedDevices, &SimulatedDevice{
			device:      device,
			geolocation: deviceGeolocation,
			// this isn't normalized, so it's not truly stepDistance units per second, but close enough
			stepDisplacementY: s.movementPerSec * math.Sin(directionRadians),
			stepDisplacementX: s.movementPerSec * math.Cos(directionRadians),
			lastUpdate:        time.Now(),
		})
	}
	return nil
}

func (s *SimulatorImpl) stepDevices(ctx context.Context) error {
	// previously, I implemented a single transaction for each device's insert
	// this wasn't slow per say, but hit the GCP IOPS limit
	// which I didn't want to pay to increase, so I'm batching the inserts now
	//
	// although it would be a better simulation to insert each device's geolocation separately
	// since this is the behavior one would expect from payloads coming from separate tracking devices

	geolocationsToInsert := []*database.DeviceGeolocation{}
	for _, device := range s.simulatedDevices {
		device.geolocation.EventTime = time.Now()

		deltaSeconds := time.Since(device.lastUpdate).Seconds()
		device.geolocation.Latitude += deltaSeconds * device.stepDisplacementY
		device.geolocation.Longitude += deltaSeconds * device.stepDisplacementX

		// switch direction if we're outside the circle
		distanceSquaredFromCenter := math.Pow(device.geolocation.Latitude-s.centerLatitude, 2) + math.Pow(device.geolocation.Longitude-s.centerLongitude, 2)
		if distanceSquaredFromCenter > math.Pow(s.radius, 2) && device.switchDirectionCooldownSteps <= 0 {
			device.stepDisplacementX *= -1
			device.stepDisplacementY *= -1
			device.switchDirectionCooldownSteps = 5
		}
		device.switchDirectionCooldownSteps--

		geolocationsToInsert = append(geolocationsToInsert, device.geolocation)
		device.lastUpdate = time.Now()
	}

	retries := 0
	var err error
	for retries < s.maxInsertRetries {
		err := s.repo.InsertMultiGeolocation(ctx, geolocationsToInsert)
		if err == nil {
			break
		}
		retries++
		time.Sleep(s.insertRetryTime)
	}
	if retries > s.maxInsertRetries {
		fmt.Printf("Failed to insert geolocations after %d retries: %v\n", retries, err)
		return err
	}

	return nil
}
