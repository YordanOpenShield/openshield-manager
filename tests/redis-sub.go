package tests

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func subscribeToChannel(client *redis.Client, channel string) {
	pubsub := client.Subscribe(ctx, channel)
	defer pubsub.Close()

	// Wait for message from the publisher (manager)
	for msg := range pubsub.Channel() {
		fmt.Printf("Received command for %s: %s\n", channel, msg.Payload)
	}
}

func main() {
	// TODO: Complete the test
}
