package appErrors

import "strings"

type CreateMissingEnvErrorInput struct {
	Location string
	Keys     []string
}

func CreateMissingEnvError(input CreateMissingEnvErrorInput) *AppError {
	return Create(CreateAppErrorInput{
		Name:     "MissingEnvError",
		Message:  "Missing required env vars: " + strings.Join(input.Keys, ", "),
		Location: input.Location,
		Level:    LevelHigh,
		Data:     map[string]any{"keys": input.Keys},
	})
}
