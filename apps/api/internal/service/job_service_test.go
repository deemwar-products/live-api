package service

import (
	"context"
	"strings"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestJobService(t *testing.T) (*JobService, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cleanup := func() {
		rdb.Close()
		mr.Close()
	}
	return NewJobService(rdb), cleanup
}

func TestJobServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	svc, cleanup := newTestJobService(t)
	defer cleanup()

	created, err := svc.CreateJob(ctx, "org_1", "process_document")
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}
	if !strings.HasPrefix(created.ID, "job_") {
		t.Fatalf("expected generated job id, got %q", created.ID)
	}

	jobs, err := svc.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs returned error: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	job, err := svc.GetJob(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetJob returned error: %v", err)
	}
	if job == nil || job.ID != created.ID {
		t.Fatalf("unexpected job: %+v", job)
	}

	missing, err := svc.GetJob(ctx, "missing")
	if err != nil {
		t.Fatalf("GetJob missing returned error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil missing job, got %+v", missing)
	}

	if err := svc.CancelJob(ctx, created.ID); err != nil {
		t.Fatalf("CancelJob returned error: %v", err)
	}

	stats, err := svc.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats returned error: %v", err)
	}
	if stats.Queued != 1 || stats.Processing != 0 || stats.Completed != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestJobServiceErrors(t *testing.T) {
	ctx := context.Background()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := NewJobService(rdb)
	mr.Close()
	rdb.Close()

	if _, err := svc.CreateJob(ctx, "org", "type"); err == nil {
		t.Fatal("expected CreateJob error")
	}
	if _, err := svc.ListJobs(ctx); err == nil {
		t.Fatal("expected ListJobs error")
	}
	if _, err := svc.GetJob(ctx, "job_1"); err == nil {
		t.Fatal("expected GetJob error")
	}
	if err := svc.CancelJob(ctx, "job_1"); err == nil {
		t.Fatal("expected CancelJob error")
	}
	if _, err := svc.GetStats(ctx); err == nil {
		t.Fatal("expected GetStats error")
	}
}
