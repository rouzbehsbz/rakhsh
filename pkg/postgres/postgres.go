package postgres

import (
	"context"
	"fmt"
	postgresDb "rakhsh/db/postgres/gen"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type Postgres struct {
	Q *postgresDb.Queries
}

func NewPostgresService(host string, port uint16, username, password, databaseName string, maxConnections int) (*Postgres, error) {
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", username, password, host, port, databaseName)

	config, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	config.MaxConns = int32(maxConnections)

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %v", err)
	}

	q := postgresDb.New(pool)

	return &Postgres{
		Q: q,
	}, nil
}
