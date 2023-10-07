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
	device           *database.Device
	geolocation      *database.DeviceGeolocation
	directionRadians float64
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
}

func New(repo database.Repo, noDevices int, centerLatitude float64, centerLongitude float64, radius float64, frequency int, movementPerSec float64) *SimulatorImpl {
	return &SimulatorImpl{
		repo:            repo,
		noDevices:       noDevices,
		centerLatitude:  centerLatitude,
		centerLongitude: centerLongitude,
		radius:          radius,
		frequency:       frequency,
		sleepTime:       time.Duration(1000/frequency) * time.Millisecond,
		movementPerSec:  movementPerSec,
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
			err := s.stepDevices(ctx)
			if err != nil {
				fmt.Printf("Simulator died. Error stepping devices: %v\n", err)
				return err
			}
			// just assume this will run fast enough to not need to compute deltas
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
	for _, device := range devices {
		directionRadians := 2 * math.Pi * rand.Float64()
		distanceAlongRadius := s.radius * rand.Float64()
		deviceGeolocation := &database.DeviceGeolocation{
			DeviceID:  device.DeviceID,
			EventTime: time.Now(),
			Latitude:  s.centerLatitude + distanceAlongRadius*math.Sin(directionRadians),
			Longitude: s.centerLongitude + distanceAlongRadius*math.Cos(directionRadians),
		}

		s.simulatedDevices = append(s.simulatedDevices, &SimulatedDevice{
			device:           device,
			geolocation:      deviceGeolocation,
			directionRadians: directionRadians,
		})
	}

	return nil
}

func (s *SimulatorImpl) stepDevices(ctx context.Context) error {
	distance := s.movementPerSec * (1.0 / float64(s.frequency))
	for _, device := range s.simulatedDevices {
		device.geolocation.EventTime = time.Now()
		device.geolocation.Latitude += distance * math.Sin(device.directionRadians)
		device.geolocation.Longitude += distance * math.Cos(device.directionRadians)

		// switch direction if we're outside the circle
		distanceSquaredFromCenter := math.Pow(device.geolocation.Latitude-s.centerLatitude, 2) + math.Pow(device.geolocation.Longitude-s.centerLongitude, 2)
		if distanceSquaredFromCenter > math.Pow(s.radius, 2) {
			device.directionRadians += math.Pi
		}

		err := s.repo.InsertGeolocation(ctx, device.geolocation)
		if err != nil {
			return err
		}
	}
	return nil
}
