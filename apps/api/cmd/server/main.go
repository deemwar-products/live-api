package main

import (
	"github.com/deemwar/live-api/apps/api/internal/server"
	"github.com/deemwar/live-api/apps/api/internal/logger"
	"os"
)

var log = logger.New("server")

func main() {
	serverInstance := server.New()
	if err := serverInstance.Start(); err != nil {
		log.Error("Server failed: %v", err)
   		os.Exit(1)
	}
}