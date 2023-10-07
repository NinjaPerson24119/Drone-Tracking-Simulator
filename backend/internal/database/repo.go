package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
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
	query := `
		INSERT INTO device.information (device_name)
		VALUES (@name)
		RETURNING device_id;
	`
	args := pgx.NamedArgs{
		"name": device.Name,
	}
	err := s.pool.QueryRow(ctx, query, args).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to insert device: %v", err)
	}
	return id, nil
}

func (s *RepoImpl) InsertGeolocation(ctx context.Context, geolocation *DeviceGeolocation) error {
	query := `
		INSERT INTO device.geolocation (device_id, event_time, latitude ,longitude)
		VALUES (@device_id, @event_time, @latitude, @longitude);
	`
	args := pgx.NamedArgs{
		"device_id":  geolocation.DeviceID,
		"event_time": geolocation.EventTime,
		"latitude":   geolocation.Latitude,
		"longitude":  geolocation.Longitude,
	}
	_, err := s.pool.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to insert geolocation: %v", err)
	}
	return nil
}

func (s *RepoImpl) GetLatestGeolocations(ctx context.Context, page int, pageSize int) ([]*DeviceGeolocation, error) {
	if page < 1 || pageSize < 1 || pageSize > 1000 {
		return nil, fmt.Errorf("invalid page or pageSize")
	}
	query := `
		SELECT d.device_id, d.event_time, d.latitude, d.longitude
		FROM device.geolocation AS d
		INNER JOIN (
			SELECT device_id, MAX(event_time) AS max_event_time
			FROM device.geolocation
			GROUP BY device_id
		) m ON m.max_event_time = d.event_time AND m.device_id = d.device_id
		ORDER BY device_id DESC
		OFFSET @offset
		LIMIT @limit;
	`
	args := pgx.NamedArgs{
		"offset": (page - 1) * pageSize,
		"limit":  pageSize,
	}
	rows, err := s.pool.Query(ctx, query, args)
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
