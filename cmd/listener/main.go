package main

import (
	"os"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/config"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/listener"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/logger"
)

func main() {
	appLogger := logger.GetInstance()
	appLogger.SetLevel(logger.LevelDebug)

	if _, err := config.GetInstance(); err != nil {
		appLogger.Error(logger.LogMessage{
			Action:  "main",
			Message: "Invalid config",
			Data:    map[string]any{"error": err.Error()},
		})
		os.Exit(1)
	}

	if err := listener.Create().Start(); err != nil {
		appLogger.Error(logger.LogMessage{
			Action:  "main",
			Message: "Listener stopped with error",
			Data:    map[string]any{"error": err.Error()},
		})
		os.Exit(1)
	}
}
