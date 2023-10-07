package database

import (
	"context"
	"fmt"
	"time"

	//"github.com/jackc/pgx/v5"
	//"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepoImpl struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, connectionURL string) (*RepoImpl, error) {
	pool, err := pgxpool.New(ctx, connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %v", err)
	}

	return &RepoImpl{
		pool: pool,
	}, nil
}

func (s *RepoImpl) InsertDevice(ctx context.Context, name string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO device.information (device_name) VALUES ($1) RETURNING device_id;
	`, name).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to insert device: %v", err)
	}
	return id, nil
}

func (s *RepoImpl) InsertGeolocation(ctx context.Context, deviceID string, eventTime time.Time, latitude float64, longitude float64) error {
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO device.geolocation (device_id, latitude, longitude) VALUES ($1, $2, $3);
	`, deviceID, eventTime, latitude, longitude)
	if err != nil {
		return fmt.Errorf("failed to insert geolocation: %v", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("failed to insert geolocation: no rows affected")
	}
	return nil
}

func (s *RepoImpl) GetLatestGeolocations(ctx context.Context, deviceID string) (string, time.Time, float64, float64, error) {
	return "", time.Time{}, 0, 0, nil
}
