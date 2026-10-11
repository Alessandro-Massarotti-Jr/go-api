package deletepostcomments

import (
	"context"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
)

type DeletePostCommentsInput struct {
	PostID entities.PostId
}

type DeletePostCommentsUseCase interface {
	Execute(ctx context.Context, input DeletePostCommentsInput) error
}

type deletePostCommentsUseCase struct {
	postCommentRepository repositories.PostCommentRepository
}

var _ DeletePostCommentsUseCase = (*deletePostCommentsUseCase)(nil)

type NewDeletePostCommentsUseCaseInput struct {
	PostCommentRepository repositories.PostCommentRepository
}

func NewDeletePostCommentsUseCase(input NewDeletePostCommentsUseCaseInput) DeletePostCommentsUseCase {
	return &deletePostCommentsUseCase{postCommentRepository: input.PostCommentRepository}
}

func (deletePostCommentsUseCase *deletePostCommentsUseCase) Execute(ctx context.Context, input DeletePostCommentsInput) error {
	return deletePostCommentsUseCase.postCommentRepository.DeleteAllByPostId(ctx, repositories.DeletePostCommentsByPostIdInput{
		PostID: input.PostID,
	})
}
