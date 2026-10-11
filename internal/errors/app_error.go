package appErrors

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

type ErrorLevel string

const (
	LevelLow    ErrorLevel = "low"
	LevelMedium ErrorLevel = "medium"
	LevelHigh   ErrorLevel = "high"
)

type AppError struct {
	Name        string         `json:"name"`
	Message     string         `json:"message"`
	Location    string         `json:"location"`
	Level       ErrorLevel     `json:"level"`
	Data        map[string]any `json:"data,omitempty"`
	HTTPStatus  int            `json:"-"`
	Stack       string         `json:"-"`
	OriginError error          `json:"-"`
}

type CreateAppErrorInput struct {
	Name        string
	Message     string
	Location    string
	Level       ErrorLevel
	Data        map[string]any
	HTTPStatus  int
	OriginError error
}

func Create(input CreateAppErrorInput) *AppError {
	if input.HTTPStatus == 0 {
		input.HTTPStatus = http.StatusInternalServerError
	}
	if input.Level == "" {
		input.Level = LevelHigh
	}
	if input.Name == "" {
		input.Name = "AppError"
	}
	if input.Message == "" {
		input.Message = "Unknown error occurred"
	}

	appError := &AppError{
		Name:        input.Name,
		Message:     input.Message,
		Location:    input.Location,
		Level:       input.Level,
		Data:        input.Data,
		HTTPStatus:  input.HTTPStatus,
		OriginError: input.OriginError,
		Stack:       string(debug.Stack()),
	}

	return appError
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s]: %s", e.Name, e.Message)
}
