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
	insertGeolocationQuery					   = `
		INSERT INTO device.geolocation (device_id, event_time, latitude ,longitude)
		VALUES (@device_id, @event_time, @latitude, @longitude);
	`
)

type RepoImpl struct {
	// this resource is thread safe
	pool          *pgxpool.Pool
	connectionURL string
}

func New(ctx context.Context, connectionURL string) (*RepoImpl, error) {
	config, err := pgxpool.ParseConfig(connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres connection url: %v", err)
	}
	config.MinConns = 10
	config.MaxConns = 100
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %v", err)
	}

	return &RepoImpl{
		pool:          pool,
		connectionURL: connectionURL,
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

func insertGeolocationNamedArgs(geolocation *DeviceGeolocation) pgx.NamedArgs {
	return pgx.NamedArgs{
		"device_id":  geolocation.DeviceID,
		"event_time": geolocation.EventTime,
		"latitude":   geolocation.Latitude,
		"longitude":  geolocation.Longitude,
	}
}

func (s *RepoImpl) InsertGeolocation(ctx context.Context, geolocation *DeviceGeolocation) error {
	args := insertGeolocationNamedArgs(geolocation)
	_, err := s.pool.Exec(ctx, insertGeolocationQuery, args)
	if err != nil {
		return fmt.Errorf("failed to insert geolocation: %v", err)
	}
	return nil
}

func (s *RepoImpl) InsertMultiGeolocation(ctx context.Context, geolocations []*DeviceGeolocation) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}

	batch := &pgx.Batch{}
	for _, geolocation := range geolocations {
		args := insertGeolocationNamedArgs(geolocation)
		batch.Queue(insertGeolocationQuery, args)
	}
	br := tx.SendBatch(ctx, batch)
	
	err = br.Close()
	if err != nil {
		return fmt.Errorf("failed to insert multi geolocation: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func 

func (s *RepoImpl) ListLatestGeolocations(ctx context.Context, paging filters.PageOptions) ([]*DeviceGeolocation, error) {
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

func (s *RepoImpl) GetMultiLatestGeolocations(ctx context.Context, deviceIDs []string) ([]*DeviceGeolocation, error) {
	query := `
	SELECT d.device_id, d.event_time, d.latitude, d.longitude, d.created, d.updated, d.deleted
	FROM device.geolocation AS d
	INNER JOIN (
		SELECT device_id, MAX(event_time) AS max_event_time
		FROM device.geolocation
		WHERE device_id = ANY(@deviceIDs) AND deleted IS NULL
		GROUP BY device_id
	) m ON m.max_event_time = d.event_time AND m.device_id = d.device_id
	WHERE d.device_id = ANY(@deviceIDs) AND d.deleted IS NULL
	ORDER BY device_id DESC
	LIMIT @lim;
`
	args := pgx.NamedArgs{
		"deviceIDs": deviceIDs,
		"lim":       len(deviceIDs),
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

	// get multi returns the same order as the input. if a device is not found, it will be nil
	geolocationsMap := map[string]*DeviceGeolocation{}
	for i := range geolocations {
		geolocationsMap[geolocations[i].DeviceID] = &geolocations[i]
	}
	ptrs := make([]*DeviceGeolocation, len(deviceIDs))
	for i := range deviceIDs {
		ptrs[i] = geolocationsMap[deviceIDs[i]]
	}
	return ptrs, nil
}

func (s *RepoImpl) ListenToGeolocationInserted(ctx context.Context, handler func(string) error) error {
	// connect directly without pool to avoid competing with other connections
	// TODO: this could really be a single connection for the entire app which multicasts to all listeners
	conn, err := pgx.Connect(ctx, s.connectionURL)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, fmt.Sprintf("LISTEN %s;", deviceGeolocationInsertedNotificationChannel))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			return fmt.Errorf("failed to wait for notification: %v", err)
		}

		deviceID := notification.Payload
		err = handler(deviceID)
		if err != nil {
			return fmt.Errorf("failed to handle notification: %v", err)
		}
	}
}
