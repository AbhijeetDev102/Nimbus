package nimbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisPublisher struct {
	client *redis.Client
}

func newRedisPublisher(client *redis.Client) ProgressPublisher {
	return &redisPublisher{client: client}
}

func (p *redisPublisher) Publish(ctx context.Context, update *ProgressUpdate) error {
	payload, err := json.Marshal(update)
	if err != nil {
		return err
	}
	channel := fmt.Sprintf("job:progress:%s", update.JobID)
	// Cache the latest progress tick for instant WebSocket retrieval on connect
	_ = p.client.Set(ctx, fmt.Sprintf("job:progress:%s:latest", update.JobID), payload, 1*time.Hour).Err()
	return p.client.Publish(ctx, channel, payload).Err()
}
