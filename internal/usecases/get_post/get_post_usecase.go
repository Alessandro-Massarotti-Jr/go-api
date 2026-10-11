package getpost

import (
	"context"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
)

type GetPostInput struct {
	ID entities.PostId
}

type GetPostUseCase interface {
	Execute(ctx context.Context, input GetPostInput) (*entities.Post, error)
}

type getPostUseCase struct {
	postRepository repositories.PostRepository
}

var _ GetPostUseCase = (*getPostUseCase)(nil)

type NewGetPostUseCaseInput struct {
	PostRepository repositories.PostRepository
}

func NewGetPostUseCase(input NewGetPostUseCaseInput) GetPostUseCase {
	return &getPostUseCase{postRepository: input.PostRepository}
}

func (getPostUseCase *getPostUseCase) Execute(ctx context.Context, input GetPostInput) (*entities.Post, error) {
	return getPostUseCase.postRepository.FindById(ctx, repositories.FindPostByIdInput{ID: input.ID})
}
