package database

import (
	"context"
	"fmt"

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

func (s *RepoImpl) InsertDevice(ctx context.Context, device *Device) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO device.information (device_name)
		VALUES ($1)
		RETURNING device_id;
	`, device.Name).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to insert device: %v", err)
	}
	return id, nil
}

func (s *RepoImpl) InsertGeolocation(ctx context.Context, geolocation *DeviceGeolocation) error {
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO device.geolocation (device_id, event_time, latitude ,longitude)
		VALUES ($1, $2, $3, $4);
	`, geolocation.DeviceID, geolocation.EventTime, geolocation.Latitude, geolocation.Longitude)
	if err != nil {
		return fmt.Errorf("failed to insert geolocation: %v", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("failed to insert geolocation: no rows affected")
	}
	return nil
}

func (s *RepoImpl) GetLatestGeolocations(ctx context.Context, page int, pageSize int) ([]*DeviceGeolocation, error) {
	if page < 1 || pageSize < 1 || pageSize > 1000 {
		return nil, fmt.Errorf("invalid page or pageSize")
	}
	rows, err := s.pool.Query(ctx, `
		SELECT d.device_id, d.event_time, d.latitude, d.longitude
		FROM device.geolocation AS d
		INNER JOIN (
			SELECT device_id, MAX(event_time) AS max_event_time
			FROM device.geolocation
			GROUP BY device_id
		) m ON m.device_id = d.device_id
		ORDER BY device_id DESC
		OFFSET $1
		LIMIT $2;
	`, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest geolocations: %v", err)
	}
	defer rows.Close()

	var geolocations []*DeviceGeolocation
	for rows.Next() {
		var geolocation DeviceGeolocation
		err := rows.Scan(
			&geolocation.DeviceID,
			&geolocation.EventTime,
			&geolocation.Latitude,
			&geolocation.Longitude,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a geolocation: %v", err)
		}
		geolocations = append(geolocations, &geolocation)
	}

	return geolocations, nil
}
