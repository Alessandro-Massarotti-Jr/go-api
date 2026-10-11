package appErrors

import "net/http"

type CreateInvalidCursorErrorInput struct {
	Location    string
	OriginError error
}

func CreateInvalidCursorError(input CreateInvalidCursorErrorInput) *AppError {
	return Create(CreateAppErrorInput{
		Name:        "InvalidCursorError",
		Message:     "Invalid pagination cursor",
		Location:    input.Location,
		Level:       LevelLow,
		HTTPStatus:  http.StatusBadRequest,
		OriginError: input.OriginError,
	})
}
