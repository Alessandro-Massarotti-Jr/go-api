package appErrors

import "net/http"

type CreatePostAlreadyExistsErrorInput struct {
	Location    string
	ID          string
	OriginError error
}

func CreatePostAlreadyExistsError(input CreatePostAlreadyExistsErrorInput) *AppError {
	return Create(CreateAppErrorInput{
		Name:        "PostAlreadyExistsError",
		Message:     "Post already exists",
		Location:    input.Location,
		Level:       LevelLow,
		Data:        map[string]any{"id": input.ID},
		HTTPStatus:  http.StatusConflict,
		OriginError: input.OriginError,
	})
}
