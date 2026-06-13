package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Message is a decoded Redis Stream entry.
type Message struct {
	StreamID string // Redis message ID — used for XACK
	JobID    string
	Type     string
}

// Consumer reads and acknowledges messages from a Redis Stream consumer group.
type Consumer struct {
	client *redis.Client
}

// NewConsumer returns a Consumer backed by the given Redis client.
func NewConsumer(client *redis.Client) *Consumer {
	return &Consumer{client: client}
}

// CreateGroup creates a consumer group on the stream, starting from the
// beginning ("0"). If the group already exists the call is a no-op.
func (c *Consumer) CreateGroup(ctx context.Context, stream, group string) error {
	err := c.client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("create group %s on %s: %w", group, stream, err)
	}
	return nil
}

// Read blocks for up to blockDur waiting for new messages on the stream.
// Returns nil slice (no error) when the block timeout expires with no messages.
func (c *Consumer) Read(ctx context.Context, stream, group, consumer string, blockDur time.Duration) ([]Message, error) {
	result, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    10,
		Block:    blockDur,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("read from %s: %w", stream, err)
	}

	var msgs []Message
	for _, s := range result {
		for _, m := range s.Messages {
			jobID, _ := m.Values["job_id"].(string)
			jobType, _ := m.Values["type"].(string)
			msgs = append(msgs, Message{
				StreamID: m.ID,
				JobID:    jobID,
				Type:     jobType,
			})
		}
	}
	return msgs, nil
}

// Ack acknowledges one or more messages, removing them from the PEL.
func (c *Consumer) Ack(ctx context.Context, stream, group string, ids ...string) error {
	if err := c.client.XAck(ctx, stream, group, ids...).Err(); err != nil {
		return fmt.Errorf("ack %v on %s: %w", ids, stream, err)
	}
	return nil
}
