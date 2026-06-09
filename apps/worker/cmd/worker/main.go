package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"github.com/deemwar/live-api-poc/apps/worker/internal/queue/processor"
)

type workerProcessor interface {
	EnsureGroup(ctx context.Context) error
	Run(ctx context.Context)
}

var (
	newRedisClient = func(addr string) *redis.Client { return redis.NewClient(&redis.Options{Addr: addr}) }
	newProcessor   = func(rdb *redis.Client) workerProcessor { return processor.New(rdb) }
	notifySignals  = signal.Notify
	logFatalf      = log.Fatalf
	logPrintln     = log.Println
)

func redisAddr() string {
	return fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", "localhost"), getEnv("REDIS_PORT", "6379"))
}

func run(sig chan os.Signal) error {
	rdb := newRedisClient(redisAddr())
	proc := newProcessor(rdb)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := proc.EnsureGroup(ctx); err != nil {
		return err
	}

	notifySignals(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sig
		logPrintln("shutting down worker")
		cancel()
	}()

	logPrintln("worker started, waiting for jobs...")
	proc.Run(ctx)
	return nil
}

func main() {
	sig := make(chan os.Signal, 1)
	if err := run(sig); err != nil {
		logFatalf("ensure group: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
