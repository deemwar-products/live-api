package server

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/deemwar/live-api/apps/api/internal/config"
	"github.com/deemwar/live-api/apps/api/internal/db"
	"github.com/deemwar/live-api/apps/api/internal/handlers"
	"github.com/deemwar/live-api/apps/api/internal/live"
	"github.com/deemwar/live-api/apps/api/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/genai"
)

var log = logger.New("server")

type Server struct {
	engine *gin.Engine
	address string
	dbConn *sql.DB
	migrate func(*sql.DB) error
	liveHandler *live.Handler
	documentsHandler *handlers.DocumentsHandler
	queueStatusHandler *handlers.QueueStatusHandler
}

func New() *Server {
	cfg := config.Load()
	return NewWithDeps(db.OpenReadWrite, func(conn *sql.DB) error {
		return db.Migrate(conn, cfg.MigrationsPath)
	})
}

func NewWithDeps(opener func(string) (*sql.DB, error), migrator func(*sql.DB) error) *Server {
	cfg := config.Load()
	log.Info("Creating server instance")

	dbConn, err := opener(cfg.DBPath)
	if err != nil {
		log.Error("Failed to open database: %v", err)
		panic(err)
	}

	// Redis client used by both the documents handler (XADD) and queue status (XLEN).
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB: cfg.RedisDB,
	})
	queueAdapter := &handlers.RedisClient{C: redisClient}

	sessions := live.NewSessionManager(cfg.MaxSessions)

	var liveHandler *live.Handler
	if cfg.GeminiAPIKey != "" {
		genaiClient, err := genai.NewClient(context.Background(), &genai.ClientConfig{
			APIKey: cfg.GeminiAPIKey,
			HTTPOptions: genai.HTTPOptions{
				APIVersion: "v1alpha",
			},
		})
		if err != nil {
			log.Error("Failed to create Gemini client: %v", err)
		} else {
			liveHandler = live.NewHandler(cfg, sessions, genaiClient)
		}
	}

	docsHandler := handlers.NewDocumentsHandler(dbConn, queueAdapter, cfg.DocumentsDir, cfg.UploadMaxBytes)
	statusHandler := handlers.NewQueueStatusHandler(dbConn, queueAdapter, handlers.RagStreamName)

	return &Server{
		engine: gin.New(),
		address: fmt.Sprintf(":%s", cfg.Port),
		dbConn: dbConn,
		migrate: migrator,
		liveHandler: liveHandler,
		documentsHandler: docsHandler,
		queueStatusHandler: statusHandler,
	}
}

func (server *Server) Start() error {
	log.Info("Running database migrations")
	if err := server.migrate(server.dbConn); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	server.addMiddleware()
	server.registerRoutes()
	log.Info("Starting server on %s", server.address)
	return server.engine.Run(server.address)
}

func (server *Server) Engine() *gin.Engine {
	return server.engine
}
