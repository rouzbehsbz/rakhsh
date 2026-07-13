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
