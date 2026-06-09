package main

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestGetEnv(t *testing.T) {
	t.Setenv("REDIS_HOST", "redis")
	if got := getEnv("REDIS_HOST", "localhost"); got != "redis" {
		t.Fatalf("expected redis, got %q", got)
	}
	t.Setenv("REDIS_PORT", "")
	if got := getEnv("REDIS_PORT", "6379"); got != "6379" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestRedisAddr(t *testing.T) {
	t.Setenv("REDIS_HOST", "redis")
	t.Setenv("REDIS_PORT", "6380")
	if got := redisAddr(); got != "redis:6380" {
		t.Fatalf("unexpected redis addr %q", got)
	}
}

type fakeProcessor struct {
	ensureGroupFn func(context.Context) error
	runFn         func(context.Context)
}

func (f fakeProcessor) EnsureGroup(ctx context.Context) error {
	return f.ensureGroupFn(ctx)
}

func (f fakeProcessor) Run(ctx context.Context) {
	f.runFn(ctx)
}

func TestRun(t *testing.T) {
	origNewRedisClient := newRedisClient
	origNewProcessor := newProcessor
	origNotifySignals := notifySignals
	origLogPrintln := logPrintln
	defer func() {
		newRedisClient = origNewRedisClient
		newProcessor = origNewProcessor
		notifySignals = origNotifySignals
		logPrintln = origLogPrintln
	}()

	newRedisClient = func(addr string) *redis.Client {
		if addr != "redis:6380" {
			t.Fatalf("unexpected addr %q", addr)
		}
		return redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	}
	notifySignals = func(c chan<- os.Signal, _ ...os.Signal) {}
	logPrintln = func(...any) {}

	runStarted := make(chan struct{})
	newProcessor = func(*redis.Client) workerProcessor {
		return fakeProcessor{
			ensureGroupFn: func(context.Context) error { return nil },
			runFn: func(ctx context.Context) {
				close(runStarted)
				<-ctx.Done()
			},
		}
	}

	t.Setenv("REDIS_HOST", "redis")
	t.Setenv("REDIS_PORT", "6380")

	sig := make(chan os.Signal, 1)
	done := make(chan error, 1)
	go func() {
		done <- run(sig)
	}()

	<-runStarted
	sig <- syscall.SIGINT

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not exit after signal")
	}
}

func TestRunEnsureGroupError(t *testing.T) {
	origNewRedisClient := newRedisClient
	origNewProcessor := newProcessor
	defer func() {
		newRedisClient = origNewRedisClient
		newProcessor = origNewProcessor
	}()

	newRedisClient = func(string) *redis.Client { return redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}) }
	newProcessor = func(*redis.Client) workerProcessor {
		return fakeProcessor{
			ensureGroupFn: func(context.Context) error { return errors.New("boom") },
			runFn:         func(context.Context) {},
		}
	}

	if err := run(make(chan os.Signal, 1)); err == nil {
		t.Fatal("expected run error")
	}
}

func TestMainFatalPath(t *testing.T) {
	origLogFatalf := logFatalf
	origNewRedisClient := newRedisClient
	origNewProcessor := newProcessor
	defer func() {
		logFatalf = origLogFatalf
		newRedisClient = origNewRedisClient
		newProcessor = origNewProcessor
	}()

	newRedisClient = func(string) *redis.Client { return redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}) }
	newProcessor = func(*redis.Client) workerProcessor {
		return fakeProcessor{
			ensureGroupFn: func(context.Context) error { return errors.New("boom") },
			runFn:         func(context.Context) {},
		}
	}

	called := false
	logFatalf = func(string, ...any) { called = true }

	main()

	if !called {
		t.Fatal("expected main to call logFatalf")
	}
}
