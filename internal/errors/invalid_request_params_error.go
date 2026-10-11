package appErrors

import "net/http"

type CreateInvalidRequestParamsErrorInput struct {
	Location    string
	OriginError error
}

func CreateInvalidRequestParamsError(input CreateInvalidRequestParamsErrorInput) *AppError {
	data := map[string]any{}
	if input.OriginError != nil {
		data["reason"] = input.OriginError.Error()
	}

	return Create(CreateAppErrorInput{
		Name:        "InvalidRequestParamsError",
		Message:     "Invalid request params",
		Location:    input.Location,
		Level:       LevelLow,
		Data:        data,
		HTTPStatus:  http.StatusBadRequest,
		OriginError: input.OriginError,
	})
}
