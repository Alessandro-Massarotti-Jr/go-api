package appErrors

type CreateEventProviderErrorInput struct {
	Location    string
	Message     string
	Topic       string
	OriginError error
}

func CreateEventProviderError(input CreateEventProviderErrorInput) *AppError {
	data := map[string]any{}
	if input.Topic != "" {
		data["topic"] = input.Topic
	}

	return Create(CreateAppErrorInput{
		Name:        "EventProviderError",
		Message:     input.Message,
		Location:    input.Location,
		Level:       LevelHigh,
		Data:        data,
		OriginError: input.OriginError,
	})
}
