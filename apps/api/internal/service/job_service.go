package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const jobStream = "job_stream"
const eventStream = "job_events"

type Job struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	OrgID     string    `json:"org_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type JobStats struct {
	Queued     int64 `json:"queued"`
	Processing int64 `json:"processing"`
	Completed  int64 `json:"completed"`
}

type JobService struct {
	rdb *redis.Client
}

func NewJobService(rdb *redis.Client) *JobService {
	return &JobService{rdb: rdb}
}

func (s *JobService) CreateJob(ctx context.Context, orgID, jobType string) (*Job, error) {
	job := &Job{
		ID:        "job_" + uuid.New().String(),
		Type:      jobType,
		OrgID:     orgID,
		Status:    "queued",
		CreatedAt: time.Now().UTC(),
	}
	_, err := s.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: jobStream,
		Values: map[string]interface{}{
			"id":         job.ID,
			"type":       job.Type,
			"org_id":     job.OrgID,
			"status":     job.Status,
			"created_at": job.CreatedAt.Format(time.RFC3339),
		},
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("xadd: %w", err)
	}
	return job, nil
}

func (s *JobService) ListJobs(ctx context.Context) ([]Job, error) {
	msgs, err := s.rdb.XRange(ctx, jobStream, "-", "+").Result()
	if err != nil {
		return nil, fmt.Errorf("xrange: %w", err)
	}
	jobs := make([]Job, 0, len(msgs))
	for _, m := range msgs {
		jobs = append(jobs, jobFromValues(m.Values))
	}
	return jobs, nil
}

func (s *JobService) GetJob(ctx context.Context, id string) (*Job, error) {
	msgs, err := s.rdb.XRange(ctx, jobStream, "-", "+").Result()
	if err != nil {
		return nil, fmt.Errorf("xrange: %w", err)
	}
	for _, m := range msgs {
		if fmt.Sprintf("%v", m.Values["id"]) == id {
			j := jobFromValues(m.Values)
			return &j, nil
		}
	}
	return nil, nil
}

func (s *JobService) CancelJob(ctx context.Context, id string) error {
	_, err := s.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: eventStream,
		Values: map[string]interface{}{
			"job_id": id,
			"status": "cancelled",
		},
	}).Result()
	return err
}

func (s *JobService) GetStats(ctx context.Context) (*JobStats, error) {
	count, err := s.rdb.XLen(ctx, jobStream).Result()
	if err != nil {
		return nil, err
	}
	return &JobStats{Queued: count}, nil
}

func jobFromValues(v map[string]interface{}) Job {
	createdAt, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", v["created_at"]))
	return Job{
		ID:        fmt.Sprintf("%v", v["id"]),
		Type:      fmt.Sprintf("%v", v["type"]),
		OrgID:     fmt.Sprintf("%v", v["org_id"]),
		Status:    fmt.Sprintf("%v", v["status"]),
		CreatedAt: createdAt,
	}
}
