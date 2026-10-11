package appErrors

type CreatePostRepositoryErrorInput struct {
	Location    string
	Message     string
	OriginError error
}

func CreatePostRepositoryError(input CreatePostRepositoryErrorInput) *AppError {
	return Create(CreateAppErrorInput{
		Name:        "PostRepositoryError",
		Message:     input.Message,
		Location:    input.Location,
		Level:       LevelHigh,
		OriginError: input.OriginError,
	})
}
