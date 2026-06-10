package server

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func (server *Server) registerRoutes() {
	server.engine.GET("/health", server.health)
}

func (server *Server) health(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}