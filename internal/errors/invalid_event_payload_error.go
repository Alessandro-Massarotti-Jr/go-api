package appErrors

type CreateInvalidEventPayloadErrorInput struct {
	Location    string
	Topic       string
	OriginError error
}

func CreateInvalidEventPayloadError(input CreateInvalidEventPayloadErrorInput) *AppError {
	data := map[string]any{"topic": input.Topic}
	if input.OriginError != nil {
		data["reason"] = input.OriginError.Error()
	}

	return Create(CreateAppErrorInput{
		Name:        "InvalidEventPayloadError",
		Message:     "Invalid event payload",
		Location:    input.Location,
		Level:       LevelMedium,
		Data:        data,
		OriginError: input.OriginError,
	})
}
