package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisProgressPublisher struct {
	client *redis.Client
}

func NewRedisProgressPublisher(addr string) *RedisProgressPublisher {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &RedisProgressPublisher{
		client: client,
	}
}

func (p *RedisProgressPublisher) Publish(ctx context.Context, update *domain.ProgressUpdate) error {

	payload, err := json.Marshal(update)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("job:progress:%s", update.JobID)
	return p.client.Publish(ctx, channel, payload).Err()
}
