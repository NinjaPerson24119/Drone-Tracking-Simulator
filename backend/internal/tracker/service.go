package tracker

import (
	"context"
)

type ServiceImpl struct {
	// repo
}

func NewServiceImpl() ServiceImpl {
	return &ServiceImpl{}
}

func (s *ServiceImpl) InsertDevice(ctx context.Context, name string) (id string, error) {
	return "", nil
}

func (s *ServiceImpl) AddGeolocation(ctx context.Context, deviceID string, latitude float64, longitude float64) error {
	return nil
}
