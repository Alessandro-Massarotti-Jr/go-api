package main

import (
	"os"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/config"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/logger"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/server"
)

// @title			Go API
// @version		1.0
// @description	Estudo de estruturar um projeto com GoLang.
// @BasePath		/
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

	if err := server.Create().Start(); err != nil {
		appLogger.Error(logger.LogMessage{
			Action:  "main",
			Message: "Server stopped with error",
			Data:    map[string]any{"error": err.Error()},
		})
		os.Exit(1)
	}
}
