package main

import (
	"fmt"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/deemwar/live-api-poc/apps/api/internal/config"
	"github.com/deemwar/live-api-poc/apps/api/internal/db"
	"github.com/deemwar/live-api-poc/apps/api/internal/handler"
	"github.com/deemwar/live-api-poc/apps/api/internal/middleware"
	"github.com/deemwar/live-api-poc/apps/api/internal/service"
)

var (
	loadConfig    = config.Load
	openDatabase  = func(path string) (io.Closer, error) { return db.Open(path) }
	newRedis      = func(addr string) *redis.Client { return redis.NewClient(&redis.Options{Addr: addr}) }
	runHTTPServer = func(r *gin.Engine, addr string) error { return r.Run(addr) }
	logFatalf     = log.Fatalf
	logPrintf     = log.Printf
)

func newRouter(jobHandler *handler.JobHandler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", handler.Health)

	v1 := r.Group("/api/v1")
	{
		jobs := v1.Group("/queue/jobs")
		jobs.GET("", jobHandler.List)
		jobs.GET("/:id", jobHandler.Get)
		jobs.POST("", jobHandler.Create)
		jobs.DELETE("/:id", jobHandler.Cancel)
		v1.GET("/queue/stats", jobHandler.Stats)
	}

	return r
}

func run() error {
	cfg := loadConfig()

	database, err := openDatabase(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer database.Close()

	rdb := newRedis(fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort))
	jobSvc := service.NewJobService(rdb)
	jobHandler := handler.NewJobHandler(jobSvc)
	r := newRouter(jobHandler)

	logPrintf("api listening on :%s", cfg.Port)
	return runHTTPServer(r, ":"+cfg.Port)
}

func main() {
	if err := run(); err != nil {
		logFatalf("run: %v", err)
	}
}
