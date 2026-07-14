package redis

import (
	"context"
	"encoding/json"
	"rakhsh/internal/common"
	"time"

	goRedis "github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *goRedis.Client
}

func NewRedis(url string, password string, maxConnections int) (*Redis, error) {
	client := goRedis.NewClient(&goRedis.Options{
		Addr:     url,
		Password: password,
		PoolSize: maxConnections,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &Redis{
		Client: client,
	}, nil
}

func (r *Redis) GetJson(ctx context.Context, key string, dest any) error {
	res, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		if err == goRedis.Nil {
			return common.ErrNotFound
		}

		return err
	}

	return json.Unmarshal(res, dest)
}

func (r *Redis) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.Client.Set(ctx, key, data, ttl).Err()
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

func (r *Redis) GetInt(ctx context.Context, key string) (int, error) {
	n, err := r.Client.Get(ctx, key).Int()
	if err != nil {
		if err == goRedis.Nil {
			return 0, common.ErrNotFound
		}

		return 0, err
	}

	return n, nil
}

func (r *Redis) GetInt64(ctx context.Context, key string) (uint64, error) {
	n, err := r.Client.Get(ctx, key).Int64()
	if err != nil {
		if err == goRedis.Nil {
			return 0, common.ErrNotFound
		}

		return 0, err
	}

	return uint64(n), nil
}

func (r *Redis) AddInt(ctx context.Context, key string, amount int64) (int64, error) {
	value, err := r.Client.IncrBy(ctx, key, amount).Result()
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (r *Redis) DeleteMany(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}
