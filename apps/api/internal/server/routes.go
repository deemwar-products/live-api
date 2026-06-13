package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) registerRoutes() {
	server.engine.GET("/health", server.health)
	server.engine.GET("/ws", server.handleWS)

	api := server.engine.Group("/api/v1")
	api.POST("/documents", server.documentsHandler.Upload)
	api.GET("/documents", server.documentsHandler.List)
	api.GET("/documents/:id", server.documentsHandler.Get)
	api.POST("/documents/:id/requeue", server.documentsHandler.Requeue)
	api.GET("/queue/status", server.queueStatusHandler.Status)
}

func (server *Server) health(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (server *Server) handleWS(c *gin.Context) {
	if server.liveHandler == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "live session not configured — set GEMINI_API_KEY"})
		return
	}
	server.liveHandler.ServeHTTP(c)
}
