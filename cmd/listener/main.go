package main

import (
	"os"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/config"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/listener"
)

func main() {
	logger := config.GetInstance()
	logger.SetLevel(config.LevelDebug)

	if err := listener.Create().Start(); err != nil {
		logger.Error(config.LogMessage{
			Action:  "main",
			Message: "Listener stopped with error",
			Data:    map[string]any{"error": err.Error()},
		})
		os.Exit(1)
	}
}
