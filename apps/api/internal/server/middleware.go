package server

import "github.com/gin-gonic/gin"

func (server *Server) addMiddleware() {
	server.engine.Use(gin.Recovery())
}