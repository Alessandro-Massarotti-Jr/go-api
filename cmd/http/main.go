package main

import (
	"os"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/config"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/server"
)

func main() {
	logger := config.GetInstance()
	logger.SetLevel(config.LevelDebug)

	if err := server.Create().Start(); err != nil {
		logger.Error(config.LogMessage{
			Action:  "main",
			Message: "Server stopped with error",
			Data:    map[string]any{"error": err.Error()},
		})
		os.Exit(1)
	}
}
