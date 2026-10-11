package appErrors

import "net/http"

type CreatePostNotFoundErrorInput struct {
	Location string
	ID       string
}

func CreatePostNotFoundError(input CreatePostNotFoundErrorInput) *AppError {
	return Create(CreateAppErrorInput{
		Name:       "PostNotFoundError",
		Message:    "Post not found",
		Location:   input.Location,
		Level:      LevelLow,
		Data:       map[string]any{"id": input.ID},
		HTTPStatus: http.StatusNotFound,
	})
}
