package main

import (
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/config"
)

func main() {
	logger := config.GetInstance()
	logger.SetLevel(config.LevelDebug)
	logger.Debug(config.LogMessage{
		Action:  "main",
		Message: "Debug level set",
		Data: map[string]any{
			"example_key": "example_value",
			"another_key": 123,
		},
	})
	logger.Info(config.LogMessage{
		Action:  "main",
		Message: "Info level message",
		Data:    nil,
	})
	logger.Warn(config.LogMessage{
		Action:  "main",
		Message: "Warn level message",
		Data:    nil,
	})
	logger.Error(config.LogMessage{
		Action:  "main",
		Message: "Error level message",
		Data:    nil,
	})
}
