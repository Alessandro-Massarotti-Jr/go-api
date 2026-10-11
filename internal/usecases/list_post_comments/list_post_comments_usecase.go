package listpostcomments

import (
	"context"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
)

type ListPostCommentsInput struct {
	PostID entities.PostId
	Cursor *string
	Limit  int
}

type ListPostCommentsUseCase interface {
	Execute(ctx context.Context, input ListPostCommentsInput) (*repositories.PaginatedPostComments, error)
}

type listPostCommentsUseCase struct {
	postRepository        repositories.PostRepository
	postCommentRepository repositories.PostCommentRepository
}

var _ ListPostCommentsUseCase = (*listPostCommentsUseCase)(nil)

type NewListPostCommentsUseCaseInput struct {
	PostRepository        repositories.PostRepository
	PostCommentRepository repositories.PostCommentRepository
}

func NewListPostCommentsUseCase(input NewListPostCommentsUseCaseInput) ListPostCommentsUseCase {
	return &listPostCommentsUseCase{
		postRepository:        input.PostRepository,
		postCommentRepository: input.PostCommentRepository,
	}
}

// Execute returns the post comments, or PostNotFoundError when the post does not exist.
func (listPostCommentsUseCase *listPostCommentsUseCase) Execute(ctx context.Context, input ListPostCommentsInput) (*repositories.PaginatedPostComments, error) {
	if _, err := listPostCommentsUseCase.postRepository.FindById(ctx, repositories.FindPostByIdInput{ID: input.PostID}); err != nil {
		return nil, err
	}

	return listPostCommentsUseCase.postCommentRepository.FindAllByPostId(ctx, repositories.FindPostCommentsByPostIdInput{
		PostID: input.PostID,
		Cursor: input.Cursor,
		Limit:  input.Limit,
	})
}
