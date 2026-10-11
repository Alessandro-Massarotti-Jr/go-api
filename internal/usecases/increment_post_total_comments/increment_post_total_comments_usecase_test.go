package incrementposttotalcomments_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	incrementposttotalcomments "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/increment_post_total_comments"
)

type postRepositoryMock struct {
	IncrementTotalCommentsFunc func(ctx context.Context, input repositories.IncrementPostTotalCommentsInput) error
}

var _ repositories.PostRepository = (*postRepositoryMock)(nil)

func (postRepositoryMock *postRepositoryMock) IncrementTotalComments(ctx context.Context, input repositories.IncrementPostTotalCommentsInput) error {
	return postRepositoryMock.IncrementTotalCommentsFunc(ctx, input)
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

func (postRepositoryMock *postRepositoryMock) Delete(ctx context.Context, input repositories.DeletePostInput) error {
	panic("postRepositoryMock.Delete should not be called")
}

func TestIncrementPostTotalCommentsUseCase(test *testing.T) {

	test.Run("Execute", func(test *testing.T) {
		test.Run("Should send the given post ID to the repository", func(test *testing.T) {
			var received repositories.IncrementPostTotalCommentsInput
			incrementPostTotalCommentsUseCase := incrementposttotalcomments.NewIncrementPostTotalCommentsUseCase(incrementposttotalcomments.NewIncrementPostTotalCommentsUseCaseInput{
				PostRepository: &postRepositoryMock{
					IncrementTotalCommentsFunc: func(ctx context.Context, input repositories.IncrementPostTotalCommentsInput) error {
						received = input
						return nil
					},
				},
			})

			err := incrementPostTotalCommentsUseCase.Execute(context.Background(), incrementposttotalcomments.IncrementPostTotalCommentsInput{PostID: "post-id"})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if received.ID != "post-id" {
				test.Errorf("ID = %q, want %q", received.ID, "post-id")
			}
		})

		test.Run("Should pass the received context to the repository", func(test *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "event")

			var receivedCtx context.Context
			incrementPostTotalCommentsUseCase := incrementposttotalcomments.NewIncrementPostTotalCommentsUseCase(incrementposttotalcomments.NewIncrementPostTotalCommentsUseCaseInput{
				PostRepository: &postRepositoryMock{
					IncrementTotalCommentsFunc: func(ctx context.Context, input repositories.IncrementPostTotalCommentsInput) error {
						receivedCtx = ctx
						return nil
					},
				},
			})

			incrementPostTotalCommentsUseCase.Execute(ctx, incrementposttotalcomments.IncrementPostTotalCommentsInput{PostID: "post-id"})

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "event" {
				test.Errorf("Expected the repository to receive the caller context, got %v", receivedCtx)
			}
		})

		test.Run("Should return the repository error", func(test *testing.T) {
			repositoryError := errors.New("repository failed")
			incrementPostTotalCommentsUseCase := incrementposttotalcomments.NewIncrementPostTotalCommentsUseCase(incrementposttotalcomments.NewIncrementPostTotalCommentsUseCaseInput{
				PostRepository: &postRepositoryMock{
					IncrementTotalCommentsFunc: func(ctx context.Context, input repositories.IncrementPostTotalCommentsInput) error {
						return repositoryError
					},
				},
			})

			err := incrementPostTotalCommentsUseCase.Execute(context.Background(), incrementposttotalcomments.IncrementPostTotalCommentsInput{PostID: "post-id"})

			if !errors.Is(err, repositoryError) {
				test.Errorf("Error = %v, want %v", err, repositoryError)
			}
		})
	})

}
