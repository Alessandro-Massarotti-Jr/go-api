package deletepostcomments_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	deletepostcomments "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/delete_post_comments"
)

type postCommentRepositoryMock struct {
	DeleteAllByPostIdFunc func(ctx context.Context, input repositories.DeletePostCommentsByPostIdInput) error
}

var _ repositories.PostCommentRepository = (*postCommentRepositoryMock)(nil)

func (postCommentRepositoryMock *postCommentRepositoryMock) DeleteAllByPostId(ctx context.Context, input repositories.DeletePostCommentsByPostIdInput) error {
	return postCommentRepositoryMock.DeleteAllByPostIdFunc(ctx, input)
}

func (postCommentRepositoryMock *postCommentRepositoryMock) FindAllByPostId(ctx context.Context, input repositories.FindPostCommentsByPostIdInput) (*repositories.PaginatedPostComments, error) {
	panic("postCommentRepositoryMock.FindAllByPostId should not be called")
}

func (postCommentRepositoryMock *postCommentRepositoryMock) Create(ctx context.Context, postComment *entities.PostComment) (*entities.PostComment, error) {
	panic("postCommentRepositoryMock.Create should not be called")
}

func TestDeletePostCommentsUseCase(test *testing.T) {

	test.Run("Execute", func(test *testing.T) {
		test.Run("Should send the given post ID to the repository", func(test *testing.T) {
			var received repositories.DeletePostCommentsByPostIdInput
			deletePostCommentsUseCase := deletepostcomments.NewDeletePostCommentsUseCase(deletepostcomments.NewDeletePostCommentsUseCaseInput{
				PostCommentRepository: &postCommentRepositoryMock{
					DeleteAllByPostIdFunc: func(ctx context.Context, input repositories.DeletePostCommentsByPostIdInput) error {
						received = input
						return nil
					},
				},
			})

			err := deletePostCommentsUseCase.Execute(context.Background(), deletepostcomments.DeletePostCommentsInput{PostID: "post-id"})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if received.PostID != "post-id" {
				test.Errorf("PostID = %q, want %q", received.PostID, "post-id")
			}
		})

		test.Run("Should pass the received context to the repository", func(test *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "event")

			var receivedCtx context.Context
			deletePostCommentsUseCase := deletepostcomments.NewDeletePostCommentsUseCase(deletepostcomments.NewDeletePostCommentsUseCaseInput{
				PostCommentRepository: &postCommentRepositoryMock{
					DeleteAllByPostIdFunc: func(ctx context.Context, input repositories.DeletePostCommentsByPostIdInput) error {
						receivedCtx = ctx
						return nil
					},
				},
			})

			deletePostCommentsUseCase.Execute(ctx, deletepostcomments.DeletePostCommentsInput{PostID: "post-id"})

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "event" {
				test.Errorf("Expected the repository to receive the caller context, got %v", receivedCtx)
			}
		})

		test.Run("Should return the repository error", func(test *testing.T) {
			repositoryError := errors.New("repository failed")
			deletePostCommentsUseCase := deletepostcomments.NewDeletePostCommentsUseCase(deletepostcomments.NewDeletePostCommentsUseCaseInput{
				PostCommentRepository: &postCommentRepositoryMock{
					DeleteAllByPostIdFunc: func(ctx context.Context, input repositories.DeletePostCommentsByPostIdInput) error {
						return repositoryError
					},
				},
			})

			err := deletePostCommentsUseCase.Execute(context.Background(), deletepostcomments.DeletePostCommentsInput{PostID: "post-id"})

			if !errors.Is(err, repositoryError) {
				test.Errorf("Error = %v, want %v", err, repositoryError)
			}
		})
	})

}
