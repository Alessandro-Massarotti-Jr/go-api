package appErrors

type CreatePostCommentRepositoryErrorInput struct {
	Location    string
	Message     string
	OriginError error
}

func CreatePostCommentRepositoryError(input CreatePostCommentRepositoryErrorInput) *AppError {
	return Create(CreateAppErrorInput{
		Name:        "PostCommentRepositoryError",
		Message:     input.Message,
		Location:    input.Location,
		Level:       LevelHigh,
		OriginError: input.OriginError,
	})
}
