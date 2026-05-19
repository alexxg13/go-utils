package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

func New(username, password, host, port, database string) (*Database, error) {
	dataSource := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		username, password, host, port, database)

	pool, err := pgxpool.New(context.Background(), dataSource)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		pool.Close()
		return nil, err
	}

	return &Database{
		Pool: pool,
	}, nil
}

func (d *Database) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}
