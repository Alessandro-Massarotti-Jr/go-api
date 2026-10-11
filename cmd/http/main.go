package main

import (
	"os"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/config"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/server"
)

// @title			Go API
// @version		1.0
// @description	Estudo de estruturar um projeto com GoLang.
// @BasePath		/
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
