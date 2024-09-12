package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.elastic.co/apm/module/apmpgxv5/v2"
)

func NewDatabase(uri string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}

	apmpgxv5.Instrument(config.ConnConfig)

	// Set up connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
