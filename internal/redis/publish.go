package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// Publish to Redis channel
func publish(client *redis.Client, channel string, payload any) error {
	err := client.Publish(ctx, channel, payload).Err()
	if err != nil {
		return fmt.Errorf("could not publish message: %v", err)
	}
	return nil
}

// Publish Job
func PublishJob(client *redis.Client, agentId string, job_name string) {
	publish(client, agentId, job_name)
}
