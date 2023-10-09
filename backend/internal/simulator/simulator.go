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
}

type SimulatorImpl struct {
	repo database.Repo

	simulatedDevices []*SimulatedDevice

	noDevices       int
	centerLatitude  float64
	centerLongitude float64
	radius          float64
	frequency       int
	sleepTime       time.Duration
	movementPerSec  float64

	maxRetries int
	retryTime  time.Duration
}

func New(repo database.Repo, noDevices int, centerLatitude float64, centerLongitude float64, radius float64, frequency int, movementPerSec float64) *SimulatorImpl {
	sleepTime := time.Duration(1000/frequency) * time.Millisecond
	return &SimulatorImpl{
		repo:            repo,
		noDevices:       noDevices,
		centerLatitude:  centerLatitude,
		centerLongitude: centerLongitude,
		radius:          radius,
		frequency:       frequency,
		sleepTime:       sleepTime,
		movementPerSec:  movementPerSec,
		maxRetries:      5,
		retryTime:       sleepTime / 5,
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
			retries := 0
			for {
				err := s.stepDevices(ctx)
				if err == nil {
					break
				}
				retries++
				if retries > s.maxRetries {
					fmt.Printf("Simulator died after %d retries. Error stepping devices: %v\n", retries, err)
					return err
				}
				time.Sleep(s.retryTime)
			}
			// this isn't an exact science. as long as the devices update continuously, we're good
			// - don't want to compute deltas
			// - don't want to factor retry times into the sleep time
			time.Sleep(s.sleepTime)
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
	stepDistance := s.movementPerSec * (1.0 / float64(s.frequency))
	for _, device := range devices {
		deviceGeolocation := &database.DeviceGeolocation{
			DeviceID:  device.DeviceID,
			EventTime: time.Now(),
			Latitude:  s.centerLatitude + s.radius/2*(rand.Float64()-0.5)*2,
			Longitude: s.centerLongitude + s.radius/2*(rand.Float64()-0.5)*2,
		}

		directionRadians := 2 * math.Pi * rand.Float64()
		s.simulatedDevices = append(s.simulatedDevices, &SimulatedDevice{
			device:            device,
			geolocation:       deviceGeolocation,
			stepDisplacementY: stepDistance * math.Sin(directionRadians),
			stepDisplacementX: stepDistance * math.Cos(directionRadians),
		})
	}
	return nil
}

func (s *SimulatorImpl) stepDevices(ctx context.Context) error {
	for _, device := range s.simulatedDevices {
		device.geolocation.EventTime = time.Now()

		device.geolocation.Latitude += device.stepDisplacementY
		device.geolocation.Longitude += device.stepDisplacementX

		// switch direction if we're outside the circle
		distanceSquaredFromCenter := math.Pow(device.geolocation.Latitude-s.centerLatitude, 2) + math.Pow(device.geolocation.Longitude-s.centerLongitude, 2)
		if distanceSquaredFromCenter > math.Pow(s.radius, 2) {
			device.stepDisplacementX *= -1
			device.stepDisplacementY *= -1
		}

		err := s.repo.InsertGeolocation(ctx, device.geolocation)
		if err != nil {
			return err
		}
	}
	return nil
}
