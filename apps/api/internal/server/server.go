package server

import (
 "database/sql"
 "fmt"

 "github.com/deemwar/live-api/apps/api/internal/config"
 "github.com/deemwar/live-api/apps/api/internal/db"
 "github.com/deemwar/live-api/apps/api/internal/logger"
 "github.com/gin-gonic/gin"
)

var log = logger.New("server")

type Server struct {
 engine *gin.Engine
 address string
 dbConn *sql.DB
}

func New() *Server {
 cfg := config.Load()
 log.Info("Creating server instance")

 dbConn, err := db.OpenReadWrite(cfg.DBPath)
 if err != nil {
 log.Error("Failed to open database: %v", err)
 panic(err)
 }

 return &Server{
 engine: gin.New(),
 address: fmt.Sprintf(":%s", cfg.Port),
 dbConn: dbConn,
 }
}

func (server *Server) Start() error {
 log.Info("Running database migrations")
 if err := db.Migrate(server.dbConn); err != nil {
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