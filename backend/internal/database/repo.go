package database

import (
	"context"
	"fmt"

	"github.com/NinjaPerson24119/MapProject/backend/internal/filters"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	deviceGeolocationInsertedNotificationChannel = "geolocation_inserted"
)

type RepoImpl struct {
	// this resource is thread safe
	pool *pgxpool.Pool
}

func New(ctx context.Context, connectionURL string) (*RepoImpl, error) {
	config, err := pgxpool.ParseConfig(connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres connection url: %v", err)
	}
	config.MinConns = 5
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %v", err)
	}

	return &RepoImpl{
		pool: pool,
	}, nil
}

func (s *RepoImpl) Close() {
	s.pool.Close()
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

func (s *RepoImpl) ListDevices(ctx context.Context, paging filters.PageOptions) ([]*Device, error) {
	if paging.Page < 1 || paging.PageSize < 1 || paging.PageSize > 1000 {
		return nil, fmt.Errorf("repo: invalid page or pageSize")
	}
	query := `
		SELECT device_id, device_name, created, updated, deleted
		FROM device.information
		WHERE deleted IS NULL
		ORDER BY device_id DESC
		OFFSET @offset
		LIMIT @limit;
	`
	args := pgx.NamedArgs{
		"offset": (paging.Page - 1) * paging.PageSize,
		"limit":  paging.PageSize,
	}
	rows, err := s.pool.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %v", err)
	}
	defer rows.Close()

	devices, err := pgx.CollectRows(rows, pgx.RowToStructByName[Device])
	if err != nil {
		return nil, fmt.Errorf("failed to collect devices: %v", err)
	}

	ptrs := make([]*Device, len(devices))
	for i := range devices {
		ptrs[i] = &devices[i]
	}
	return ptrs, nil
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

func (s *RepoImpl) GetLatestGeolocations(ctx context.Context, paging filters.PageOptions) ([]*DeviceGeolocation, error) {
	if paging.Page < 1 || paging.PageSize < 1 || paging.PageSize > 1000 {
		return nil, fmt.Errorf("repo: invalid page or pageSize")
	}
	query := `
		SELECT d.device_id, d.event_time, d.latitude, d.longitude, d.created, d.updated, d.deleted
		FROM device.geolocation AS d
		INNER JOIN (
			SELECT device_id, MAX(event_time) AS max_event_time
			FROM device.geolocation
			WHERE deleted IS NULL
			GROUP BY device_id
		) m ON m.max_event_time = d.event_time AND m.device_id = d.device_id
		WHERE d.deleted IS NULL
		ORDER BY device_id DESC
		OFFSET @offset
		LIMIT @limit;
	`
	args := pgx.NamedArgs{
		"offset": (paging.Page - 1) * paging.PageSize,
		"limit":  paging.PageSize,
	}
	rows, err := s.pool.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest geolocations: %v", err)
	}
	defer rows.Close()

	geolocations, err := pgx.CollectRows(rows, pgx.RowToStructByName[DeviceGeolocation])
	if err != nil {
		return nil, fmt.Errorf("failed to collect latest geolocations: %v", err)
	}

	ptrs := make([]*DeviceGeolocation, len(geolocations))
	for i := range geolocations {
		ptrs[i] = &geolocations[i]
	}
	return ptrs, nil
}

func (s *RepoImpl) GetLatestGeolocation(ctx context.Context, deviceID string) (*DeviceGeolocation, error) {
	query := `
		SELECT d.device_id, d.event_time, d.latitude, d.longitude, d.created, d.updated, d.deleted
		FROM device.geolocation AS d
		INNER JOIN (
			SELECT device_id, MAX(event_time) AS max_event_time
			FROM device.geolocation
			WHERE device_id = @deviceID AND deleted IS NULL
			GROUP BY device_id
		) m ON m.max_event_time = d.event_time AND m.device_id = d.device_id
		WHERE d.device_id = @deviceID AND d.deleted IS NULL
		ORDER BY device_id DESC
		LIMIT 1;
	`
	args := pgx.NamedArgs{
		"deviceID": deviceID,
	}
	rows, err := s.pool.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest geolocation: %v", err)
	}
	defer rows.Close()

	geolocations, err := pgx.CollectRows(rows, pgx.RowToStructByName[DeviceGeolocation])
	if err != nil {
		return nil, fmt.Errorf("failed to collect latest geolocations: %v", err)
	}
	if len(geolocations) == 0 {
		return nil, nil
	}

	return &geolocations[0], nil
}

func (s *RepoImpl) ListenToGeolocationInserted(ctx context.Context, handler func(*DeviceGeolocation) error) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, fmt.Sprintf("LISTEN %s;", deviceGeolocationInsertedNotificationChannel))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return fmt.Errorf("failed to wait for notification: %v", err)
		}

		deviceGeolocation, err := s.GetLatestGeolocation(ctx, notification.Payload)
		if err != nil {
			return fmt.Errorf("failed to get latest geolocation: %v", err)
		}
		if deviceGeolocation == nil {
			return fmt.Errorf("failed to get latest geolocation: not found")
		}

		err = handler(deviceGeolocation)
		if err != nil {
			return fmt.Errorf("failed to handle notification: %v", err)
		}
	}
}
