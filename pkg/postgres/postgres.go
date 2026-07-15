package postgres

import (
	"context"
	"fmt"
	postgresDb "rakhsh/db/postgres/gen"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type txKey struct{}

type Shard struct {
	//TODO: add support for replicas
	Master  *pgxpool.Pool
	MasterQ *postgresDb.Queries
}

type Postgres struct {
	Shards           map[int]Shard
	CelebritiesShard map[int32]int
	ShardsCount      int
}

func NewPostgresService(shardUrls []string, celebritiesShard map[int32]int, maxConnections int) (*Postgres, error) {
	ctx := context.Background()

	shards := make(map[int]Shard)

	for i, shardUrl := range shardUrls {
		masterPool, err := createPool(ctx, shardUrl, maxConnections)
		if err != nil {
			return nil, fmt.Errorf("failed to create master pool for shard %d: %w", i, err)
		}

		shards[i] = Shard{
			Master:  masterPool,
			MasterQ: postgresDb.New(masterPool),
		}
	}
	return &Postgres{
		Shards:           shards,
		ShardsCount:      len(shardUrls),
		CelebritiesShard: celebritiesShard,
	}, nil
}

func (p *Postgres) GetShard(clientId int32) Shard {
	if shardId, ok := p.CelebritiesShard[clientId]; ok {
		return p.Shards[shardId]
	}

	shardId := clientId % int32(p.ShardsCount)
	return p.Shards[int(shardId)]
}

func (p *Postgres) WithinTx(ctx context.Context, clientID int32, fn func(ctx context.Context) error) error {
	shard := p.GetShard(clientID)

	tx, err := shard.Master.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func createPool(ctx context.Context, url string, maxConnections int) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	config.MaxConns = int32(maxConnections)
	return pgxpool.NewWithConfig(ctx, config)
}
