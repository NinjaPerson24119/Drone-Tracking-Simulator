package database

import (
	"context"
	"fmt"

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

func (s *RepoImpl) InsertDevice(ctx context.Context, name string) (id string, err error) {
	//tx, err := s.pool.Query(ctx)
	//if err != nil {
	//	return "", err
	//}

	return "", nil
}

func (s *RepoImpl) AddGeolocation(ctx context.Context, deviceID string, latitude float64, longitude float64) error {
	return nil
}
