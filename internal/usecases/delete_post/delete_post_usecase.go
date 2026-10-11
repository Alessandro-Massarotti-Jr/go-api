package deletepost

import (
	"context"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
)

type DeletePostInput struct {
	ID entities.PostId
}

type DeletePostUseCase interface {
	Execute(ctx context.Context, input DeletePostInput) error
}

type deletePostUseCase struct {
	postRepository repositories.PostRepository
}

var _ DeletePostUseCase = (*deletePostUseCase)(nil)

type NewDeletePostUseCaseInput struct {
	PostRepository repositories.PostRepository
}

func NewDeletePostUseCase(input NewDeletePostUseCaseInput) DeletePostUseCase {
	return &deletePostUseCase{postRepository: input.PostRepository}
}

func (deletePostUseCase *deletePostUseCase) Execute(ctx context.Context, input DeletePostInput) error {
	return deletePostUseCase.postRepository.Delete(ctx, repositories.DeletePostInput{ID: input.ID})
}
