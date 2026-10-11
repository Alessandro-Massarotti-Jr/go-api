package updatepost

import (
	"context"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
)

// UpdatePostInput holds the post fields to change.
// Pointer fields are optional: when nil, the current value is kept.
type UpdatePostInput struct {
	ID      entities.PostId
	Title   *string
	Content *string
}

type UpdatePostUseCase interface {
	Execute(ctx context.Context, input UpdatePostInput) (*entities.Post, error)
}

type updatePostUseCase struct {
	postRepository repositories.PostRepository
}

var _ UpdatePostUseCase = (*updatePostUseCase)(nil)

type NewUpdatePostUseCaseInput struct {
	PostRepository repositories.PostRepository
}

func NewUpdatePostUseCase(input NewUpdatePostUseCaseInput) UpdatePostUseCase {
	return &updatePostUseCase{postRepository: input.PostRepository}
}

func (updatePostUseCase *updatePostUseCase) Execute(ctx context.Context, input UpdatePostInput) (*entities.Post, error) {
	return updatePostUseCase.postRepository.Update(ctx, repositories.UpdatePostInput{
		ID:      input.ID,
		Title:   input.Title,
		Content: input.Content,
	})
}
