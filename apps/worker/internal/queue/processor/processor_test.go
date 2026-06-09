package processor

import (
	"context"
	"errors"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestProcessor(t *testing.T) (*Processor, *redis.Client, func()) {
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
	return New(rdb), rdb, cleanup
}

func TestEnsureGroup(t *testing.T) {
	ctx := context.Background()
	proc, _, cleanup := newTestProcessor(t)
	defer cleanup()

	if err := proc.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup returned error: %v", err)
	}
	if err := proc.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup existing group returned error: %v", err)
	}
}

func TestEnsureGroupError(t *testing.T) {
	ctx := context.Background()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	proc := New(rdb)
	mr.Close()
	rdb.Close()

	if err := proc.EnsureGroup(ctx); err == nil {
		t.Fatal("expected EnsureGroup error")
	}
}

func TestProcess(t *testing.T) {
	ctx := context.Background()
	proc, rdb, cleanup := newTestProcessor(t)
	defer cleanup()

	if err := proc.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup returned error: %v", err)
	}

	msgID, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: jobStream,
		Values: map[string]interface{}{"id": "job_1"},
	}).Result()
	if err != nil {
		t.Fatalf("XAdd job: %v", err)
	}

	streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumerName,
		Streams:  []string{jobStream, ">"},
		Count:    1,
	}).Result()
	if err != nil {
		t.Fatalf("XReadGroup: %v", err)
	}
	if len(streams) != 1 || len(streams[0].Messages) != 1 {
		t.Fatalf("unexpected streams: %+v", streams)
	}

	proc.process(ctx, streams[0].Messages[0])

	pending, err := rdb.XPending(ctx, jobStream, consumerGroup).Result()
	if err != nil {
		t.Fatalf("XPending: %v", err)
	}
	if pending.Count != 0 {
		t.Fatalf("expected no pending messages, got %+v", pending)
	}

	entries, err := rdb.XRange(ctx, eventStream, "-", "+").Result()
	if err != nil {
		t.Fatalf("XRange events: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 events, got %d", len(entries))
	}
	if entries[0].Values["status"] != "processing" || entries[1].Values["status"] != "completed" {
		t.Fatalf("unexpected events: %+v", entries)
	}

	msgs, err := rdb.XRange(ctx, jobStream, msgID, msgID).Result()
	if err != nil {
		t.Fatalf("XRange job stream: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected job message to remain in stream, got %d", len(msgs))
	}
}

func TestRunReturnsWhenContextCancelled(t *testing.T) {
	proc, _, cleanup := newTestProcessor(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		proc.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after context cancellation")
	}
}

func TestRunContinuesAfterReadError(t *testing.T) {
	proc := New(redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		defer proc.rdb.Close()
		proc.Run(ctx)
		done <- nil
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop after cancellation")
	}
}

func TestRunProcessesMessages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	proc, rdb, cleanup := newTestProcessor(t)
	defer cleanup()

	if err := proc.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}

	if _, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: jobStream,
		Values: map[string]interface{}{"id": "job_run_1"},
	}).Result(); err != nil {
		t.Fatalf("XAdd: %v", err)
	}

	done := make(chan struct{})
	go func() {
		proc.Run(ctx)
		close(done)
	}()

	// Wait until the event stream has both events.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		entries, _ := rdb.XRange(context.Background(), eventStream, "-", "+").Result()
		if len(entries) >= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Cancel, then add a dummy message to unblock the blocking XReadGroup so
	// Run reaches the ctx.Done() check at the top of the next iteration.
	cancel()
	rdb.XAdd(context.Background(), &redis.XAddArgs{
		Stream: jobStream,
		Values: map[string]interface{}{"id": "unblock"},
	})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after cancel")
	}

	entries, err := rdb.XRange(context.Background(), eventStream, "-", "+").Result()
	if err != nil {
		t.Fatalf("XRange events: %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(entries))
	}
}

// TestRunCtxErrOnReadGroup covers the ctx.Err() != nil branch inside Run's error handler.
// A timeout context causes the blocking XReadGroup to fail with a deadline error while
// ctx.Err() is non-nil, triggering the early return.
func TestRunCtxErrOnReadGroup(t *testing.T) {
	proc, _, cleanup := newTestProcessor(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	if err := proc.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}

	done := make(chan struct{})
	go func() {
		proc.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not exit after ctx timeout")
	}
}

func TestRunSurvivesIdlePollBeforeProcessing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	proc, rdb, cleanup := newTestProcessor(t)
	defer cleanup()

	if err := proc.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}

	done := make(chan struct{})
	go func() {
		proc.Run(ctx)
		close(done)
	}()

	time.Sleep(350 * time.Millisecond)

	if _, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: jobStream,
		Values: map[string]interface{}{"id": "job_idle_1"},
	}).Result(); err != nil {
		t.Fatalf("XAdd: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		entries, _ := rdb.XRange(context.Background(), eventStream, "-", "+").Result()
		if len(entries) >= 2 {
			cancel()
			<-done
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("expected idle run loop to continue and process later message")
}
