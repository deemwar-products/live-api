package processor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const jobStream = "job_stream"
const consumerGroup = "workers"
const consumerName = "worker1"
const eventStream = "job_events"

type Processor struct {
	rdb *redis.Client
}

var readGroup = func(ctx context.Context, rdb *redis.Client, args *redis.XReadGroupArgs) ([]redis.XStream, error) {
	return rdb.XReadGroup(ctx, args).Result()
}

func New(rdb *redis.Client) *Processor {
	return &Processor{rdb: rdb}
}

func (p *Processor) EnsureGroup(ctx context.Context) error {
	err := p.rdb.XGroupCreateMkStream(ctx, jobStream, consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("xcreategroup: %w", err)
	}
	return nil
}

func (p *Processor) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		streams, err := readGroup(ctx, p.rdb, &redis.XReadGroupArgs{
			Group:    consumerGroup,
			Consumer: consumerName,
			Streams:  []string{jobStream, ">"},
			Block:    250 * time.Millisecond,
			Count:    1,
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("xreadgroup error: %v", err)
			continue
		}
		for _, stream := range streams {
			for _, msg := range stream.Messages {
				p.process(ctx, msg)
			}
		}
	}
}

func (p *Processor) process(ctx context.Context, msg redis.XMessage) {
	jobID := fmt.Sprintf("%v", msg.Values["id"])
	log.Printf("processing job %s", jobID)

	p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: eventStream,
		Values: map[string]interface{}{"job_id": jobID, "status": "processing"},
	})

	p.rdb.XAck(ctx, jobStream, consumerGroup, msg.ID)

	p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: eventStream,
		Values: map[string]interface{}{"job_id": jobID, "status": "completed"},
	})

	log.Printf("completed job %s", jobID)
}
