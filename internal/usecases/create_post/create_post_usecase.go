package createpost

import (
	"context"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
)

type CreatePostInput struct {
	Title   string
	Content string
}

type CreatePostUseCase interface {
	Execute(ctx context.Context, input CreatePostInput) (*entities.Post, error)
}

type createPostUseCase struct {
	postRepository repositories.PostRepository
}

var _ CreatePostUseCase = (*createPostUseCase)(nil)

type NewCreatePostUseCaseInput struct {
	PostRepository repositories.PostRepository
}

func NewCreatePostUseCase(input NewCreatePostUseCaseInput) CreatePostUseCase {
	return &createPostUseCase{postRepository: input.PostRepository}
}

func (createPostUseCase *createPostUseCase) Execute(ctx context.Context, input CreatePostInput) (*entities.Post, error) {
	post := entities.Create(entities.CreatePostInput{
		Title:   input.Title,
		Content: input.Content,
	})

	return createPostUseCase.postRepository.Create(ctx, post)
}
