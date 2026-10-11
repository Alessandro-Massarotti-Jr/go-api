package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type LogLevel string

const (
	LevelDebug  LogLevel = "debug"
	LevelInfo   LogLevel = "info"
	LevelWarn   LogLevel = "warn"
	LevelError  LogLevel = "error"
	LevelSilent LogLevel = "silent"
)

var logLevelWeight = map[LogLevel]int{
	LevelDebug:  10,
	LevelInfo:   20,
	LevelWarn:   30,
	LevelError:  40,
	LevelSilent: 100,
}

const isoDateFormat = "2006-01-02T15:04:05.000Z07:00"

type LogMessage struct {
	Action  string         `json:"action"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

type Logger struct {
	mu    sync.Mutex
	level LogLevel
}

var (
	instance *Logger
	once     sync.Once
)

func GetInstance() *Logger {
	once.Do(func() {
		instance = &Logger{level: LevelInfo}
	})
	return instance
}

func (logger *Logger) SetLevel(level LogLevel) {
	if _, ok := logLevelWeight[level]; !ok {
		return
	}
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.level = level
}

func (logger *Logger) Debug(message LogMessage) {
	logger.log(LevelDebug, message)
}

func (logger *Logger) Info(message LogMessage) {
	logger.log(LevelInfo, message)
}

func (logger *Logger) Warn(message LogMessage) {
	logger.log(LevelWarn, message)
}

func (logger *Logger) Error(message LogMessage) {
	logger.log(LevelError, message)
}

func (logger *Logger) log(level LogLevel, message LogMessage) {
	logger.mu.Lock()
	defer logger.mu.Unlock()

	if logLevelWeight[level] < logLevelWeight[logger.level] {
		return
	}
	payload, _ := json.Marshal(message)
	fmt.Fprintf(os.Stdout, "[%s][%s] %s\n",
		time.Now().UTC().Format(isoDateFormat),
		strings.ToUpper(string(level)),
		payload,
	)
}
