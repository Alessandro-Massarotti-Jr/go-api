package incrementposttotalcomments

import (
	"context"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
)

type IncrementPostTotalCommentsInput struct {
	PostID entities.PostId
}

type IncrementPostTotalCommentsUseCase interface {
	Execute(ctx context.Context, input IncrementPostTotalCommentsInput) error
}

type incrementPostTotalCommentsUseCase struct {
	postRepository repositories.PostRepository
}

var _ IncrementPostTotalCommentsUseCase = (*incrementPostTotalCommentsUseCase)(nil)

type NewIncrementPostTotalCommentsUseCaseInput struct {
	PostRepository repositories.PostRepository
}

func NewIncrementPostTotalCommentsUseCase(input NewIncrementPostTotalCommentsUseCaseInput) IncrementPostTotalCommentsUseCase {
	return &incrementPostTotalCommentsUseCase{postRepository: input.PostRepository}
}

func (incrementPostTotalCommentsUseCase *incrementPostTotalCommentsUseCase) Execute(ctx context.Context, input IncrementPostTotalCommentsInput) error {
	return incrementPostTotalCommentsUseCase.postRepository.IncrementTotalComments(ctx, repositories.IncrementPostTotalCommentsInput{
		ID: input.PostID,
	})
}
