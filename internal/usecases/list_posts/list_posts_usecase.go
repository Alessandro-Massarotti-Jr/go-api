package listposts

import (
	"context"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
)

type ListPostsInput struct {
	Cursor *string
	Limit  int
}

type ListPostsUseCase interface {
	Execute(ctx context.Context, input ListPostsInput) (*repositories.PaginatedPosts, error)
}

type listPostsUseCase struct {
	postRepository repositories.PostRepository
}

var _ ListPostsUseCase = (*listPostsUseCase)(nil)

type NewListPostsUseCaseInput struct {
	PostRepository repositories.PostRepository
}

func NewListPostsUseCase(input NewListPostsUseCaseInput) ListPostsUseCase {
	return &listPostsUseCase{postRepository: input.PostRepository}
}

func (listPostsUseCase *listPostsUseCase) Execute(ctx context.Context, input ListPostsInput) (*repositories.PaginatedPosts, error) {
	return listPostsUseCase.postRepository.FindAll(ctx, repositories.FindPostsInput{
		Cursor: input.Cursor,
		Limit:  input.Limit,
	})
}
