package redis

import (
	"context"
	"time"

	"github.com/laper32/goose/logging"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

func Init(options *redis.Options) *redis.Client {
	var err error
	client = redis.NewClient(options)

	// Retry
	for i := 0; i < 3; i++ {
		err = client.Ping(context.Background()).Err()
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		panic(err)
	}
	return client
}

func Close() {
	if stream != nil {
		stream.Close()
	}
	if client != nil {
		_ = client.Close()
	}
}

func Healthy() bool {
	if client != nil {
		err := client.Ping(context.Background()).Err()
		if err == nil {
			return true
		}
		logging.Error(err)
	}
	return false
}
