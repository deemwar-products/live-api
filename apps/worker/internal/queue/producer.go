package queue

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Producer enqueues jobs onto a Redis Stream.
// Redis only carries the routing key (job_id + type); the full payload lives in DuckDB.
type Producer struct {
	client *redis.Client
}

// NewProducer returns a Producer backed by the given Redis client.
func NewProducer(client *redis.Client) *Producer {
	return &Producer{client: client}
}

// Enqueue adds a message to the given stream.
func (p *Producer) Enqueue(ctx context.Context, stream, jobID, jobType string) error {
	err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]any{
			"job_id": jobID,
			"type":   jobType,
		},
	}).Err()
	if err != nil {
		return fmt.Errorf("enqueue %s to %s: %w", jobType, stream, err)
	}
	return nil
}
