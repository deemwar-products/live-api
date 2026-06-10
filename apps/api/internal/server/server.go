package server

import (
	"github.com/deemwar/live-api/apps/api/internal/logger"
	"github.com/deemwar/live-api/apps/api/internal/config"
	"github.com/gin-gonic/gin"
	"fmt"
)

var log = logger.New("server")

type Server struct {
	engine *gin.Engine
	address string
}

func New() *Server {
	config := config.Load()
	log.Info("Creating server instance")
	return &Server{
		engine: gin.New(),
		address: fmt.Sprintf(":%s", config.Port),
	}
}

func (server *Server) Start() error {
	server.addMiddleware()
	server.registerRoutes()
	log.Info("Starting server on %s", server.address)
	return server.engine.Run(server.address)
}

func (server *Server) Engine() *gin.Engine {
	return server.engine
}
