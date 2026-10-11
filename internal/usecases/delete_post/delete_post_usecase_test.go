package deletepost_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	deletepost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/delete_post"
)

type postRepositoryMock struct {
	DeleteFunc func(ctx context.Context, input repositories.DeletePostInput) error
}

var _ repositories.PostRepository = (*postRepositoryMock)(nil)

func (postRepositoryMock *postRepositoryMock) Delete(ctx context.Context, input repositories.DeletePostInput) error {
	return postRepositoryMock.DeleteFunc(ctx, input)
}

func (postRepositoryMock *postRepositoryMock) IncrementTotalComments(ctx context.Context, input repositories.IncrementPostTotalCommentsInput) error {
	panic("postRepositoryMock.IncrementTotalComments should not be called")
}

func (postRepositoryMock *postRepositoryMock) FindAll(ctx context.Context, input repositories.FindPostsInput) (*repositories.PaginatedPosts, error) {
	panic("postRepositoryMock.FindAll should not be called")
}

func (postRepositoryMock *postRepositoryMock) FindById(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
	panic("postRepositoryMock.FindById should not be called")
}

func (postRepositoryMock *postRepositoryMock) Create(ctx context.Context, post *entities.Post) (*entities.Post, error) {
	panic("postRepositoryMock.Create should not be called")
}

func (postRepositoryMock *postRepositoryMock) Update(ctx context.Context, input repositories.UpdatePostInput) (*entities.Post, error) {
	panic("postRepositoryMock.Update should not be called")
}

func TestDeletePostUseCase(test *testing.T) {

	test.Run("Execute", func(test *testing.T) {
		test.Run("Should send the given ID to the repository", func(test *testing.T) {
			var received repositories.DeletePostInput
			deletePostUseCase := deletepost.NewDeletePostUseCase(deletepost.NewDeletePostUseCaseInput{
				PostRepository: &postRepositoryMock{
					DeleteFunc: func(ctx context.Context, input repositories.DeletePostInput) error {
						received = input
						return nil
					},
				},
			})

			err := deletePostUseCase.Execute(context.Background(), deletepost.DeletePostInput{ID: "post-id"})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if received.ID != "post-id" {
				test.Errorf("ID = %q, want %q", received.ID, "post-id")
			}
		})

		test.Run("Should pass the received context to the repository", func(test *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "request")

			var receivedCtx context.Context
			deletePostUseCase := deletepost.NewDeletePostUseCase(deletepost.NewDeletePostUseCaseInput{
				PostRepository: &postRepositoryMock{
					DeleteFunc: func(ctx context.Context, input repositories.DeletePostInput) error {
						receivedCtx = ctx
						return nil
					},
				},
			})

			deletePostUseCase.Execute(ctx, deletepost.DeletePostInput{ID: "post-id"})

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the repository to receive the caller context, got %v", receivedCtx)
			}
		})

		test.Run("Should return the repository error", func(test *testing.T) {
			repositoryError := errors.New("repository failed")
			deletePostUseCase := deletepost.NewDeletePostUseCase(deletepost.NewDeletePostUseCaseInput{
				PostRepository: &postRepositoryMock{
					DeleteFunc: func(ctx context.Context, input repositories.DeletePostInput) error {
						return repositoryError
					},
				},
			})

			err := deletePostUseCase.Execute(context.Background(), deletepost.DeletePostInput{ID: "post-id"})

			if !errors.Is(err, repositoryError) {
				test.Errorf("Error = %v, want %v", err, repositoryError)
			}
		})
	})

}
