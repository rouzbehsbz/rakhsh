package cronjob

import (
	"flag"
	"log"
	"rakhsh/internal/common"
	"rakhsh/pkg/postgres"
	"rakhsh/pkg/redis"
	"time"
)

const BatchUpdatesSize = 100
const BatchUpdateInterval = 30 * time.Second

func main() {
	isDevMode := flag.Bool("dev", true, "Run program in dev mode")
	flag.Parse()

	config, err := common.NewConfig(*isDevMode)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	celebritiesShard := map[int32]int{}

	_, err = postgres.NewPostgresService([]string{
		config.PostgresShard1Url,
		config.PostgresShard2Url,
		config.PostgresShard3Url,
	}, celebritiesShard, config.PostgresMaxConnections)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}

	_, err = redis.NewRedis(config.RedisUrl, config.RedisPassword, config.RedisMaxConnections)
	if err != nil {
		log.Fatalf("failed to init redis: %v", err)
	}
}
