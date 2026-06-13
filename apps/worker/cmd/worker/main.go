package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"github.com/deemwar/live-api/apps/worker/internal/config"
	"github.com/deemwar/live-api/apps/worker/internal/queue"
	"github.com/deemwar/live-api/apps/worker/internal/rag/chunking"
	"github.com/deemwar/live-api/apps/worker/internal/rag/embedding"
	"github.com/deemwar/live-api/apps/worker/internal/rag/fetcher"
	"github.com/deemwar/live-api/apps/worker/internal/rag/store"
	"github.com/deemwar/live-api/apps/worker/internal/worker"
	"github.com/deemwar/live-api/apps/worker/internal/worker/processors"
	_ "github.com/marcboeker/go-duckdb"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config invalid: %v", err)
	}
	if err := cfg.EnsureDBDir(); err != nil {
		log.Fatalf("ensure db dir: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Single DuckDB read-write connection — worker is the sole writer.
	db, err := sql.Open("duckdb", cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	relDB, vecDB := store.NewDuckDBStoreWithDB(db)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	gemini, err := embedding.NewGeminiClient(ctx, cfg.GeminiAPIKey, cfg.EmbeddingModel)
	if err != nil {
		log.Fatalf("gemini client: %v", err)
	}

	chunker, err := chunking.New(chunking.Options{
		TargetChildTokens: cfg.TargetChildTokens,
		OverlapTokens:     cfg.OverlapTokens,
	})
	if err != nil {
		log.Fatalf("chunker: %v", err)
	}
	chunker.SetEmbedder(gemini)

	ragProc := processors.NewRAGProcessor(db, relDB, vecDB, chunker, fetcher.NewLocalFetcher(cfg.DocumentsDir))
	writeProc := processors.NewDBWriteProcessor(db)

	consumer := queue.NewConsumer(redisClient)

	w := worker.New(cfg, consumer, db, ragProc, writeProc)

	log.Printf("worker starting — rag: %s (%s) | writes: %s (%s)",
		cfg.RagStreamName, cfg.RagGroup,
		cfg.WriteStreamName, cfg.WriteGroup,
	)

	if err := w.Run(ctx); err != nil {
		log.Fatalf("worker: %v", err)
	}

	log.Println("worker stopped")
}
