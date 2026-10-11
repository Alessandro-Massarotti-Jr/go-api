package updatepost_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	updatepost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/update_post"
)

type postRepositoryMock struct {
	UpdateFunc func(ctx context.Context, input repositories.UpdatePostInput) (*entities.Post, error)
}

var _ repositories.PostRepository = (*postRepositoryMock)(nil)

func (postRepositoryMock *postRepositoryMock) Update(ctx context.Context, input repositories.UpdatePostInput) (*entities.Post, error) {
	return postRepositoryMock.UpdateFunc(ctx, input)
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

func (postRepositoryMock *postRepositoryMock) Delete(ctx context.Context, input repositories.DeletePostInput) error {
	panic("postRepositoryMock.Delete should not be called")
}

func (postRepositoryMock *postRepositoryMock) IncrementTotalComments(ctx context.Context, input repositories.IncrementPostTotalCommentsInput) error {
	panic("postRepositoryMock.IncrementTotalComments should not be called")
}

func TestUpdatePostUseCase(test *testing.T) {

	test.Run("Execute", func(test *testing.T) {
		test.Run("Should send the ID, title and content to the repository", func(test *testing.T) {
			var received repositories.UpdatePostInput
			updatePostUseCase := updatepost.NewUpdatePostUseCase(updatepost.NewUpdatePostUseCaseInput{
				PostRepository: &postRepositoryMock{
					UpdateFunc: func(ctx context.Context, input repositories.UpdatePostInput) (*entities.Post, error) {
						received = input
						return &entities.Post{}, nil
					},
				},
			})
			title := "New title"
			content := "New content"

			_, err := updatePostUseCase.Execute(context.Background(), updatepost.UpdatePostInput{ID: "post-id", Title: &title, Content: &content})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if received.ID != "post-id" {
				test.Errorf("ID = %q, want %q", received.ID, "post-id")
			}
			if received.Title == nil || *received.Title != title {
				test.Errorf("Title = %v, want %q", received.Title, title)
			}
			if received.Content == nil || *received.Content != content {
				test.Errorf("Content = %v, want %q", received.Content, content)
			}
		})

		test.Run("Should send nil fields to the repository when they are not given", func(test *testing.T) {
			var received repositories.UpdatePostInput
			updatePostUseCase := updatepost.NewUpdatePostUseCase(updatepost.NewUpdatePostUseCaseInput{
				PostRepository: &postRepositoryMock{
					UpdateFunc: func(ctx context.Context, input repositories.UpdatePostInput) (*entities.Post, error) {
						received = input
						return &entities.Post{}, nil
					},
				},
			})
			title := "New title"

			updatePostUseCase.Execute(context.Background(), updatepost.UpdatePostInput{ID: "post-id", Title: &title})

			if received.Content != nil {
				test.Errorf("Content = %v, want nil", received.Content)
			}
		})

		test.Run("Should pass the received context to the repository", func(test *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "request")

			var receivedCtx context.Context
			updatePostUseCase := updatepost.NewUpdatePostUseCase(updatepost.NewUpdatePostUseCaseInput{
				PostRepository: &postRepositoryMock{
					UpdateFunc: func(ctx context.Context, input repositories.UpdatePostInput) (*entities.Post, error) {
						receivedCtx = ctx
						return &entities.Post{}, nil
					},
				},
			})

			updatePostUseCase.Execute(ctx, updatepost.UpdatePostInput{ID: "post-id"})

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the repository to receive the caller context, got %v", receivedCtx)
			}
		})

		test.Run("Should return the post returned by the repository", func(test *testing.T) {
			persisted := &entities.Post{ID: "post-id", Title: "persisted"}
			updatePostUseCase := updatepost.NewUpdatePostUseCase(updatepost.NewUpdatePostUseCaseInput{
				PostRepository: &postRepositoryMock{
					UpdateFunc: func(ctx context.Context, input repositories.UpdatePostInput) (*entities.Post, error) {
						return persisted, nil
					},
				},
			})

			post, err := updatePostUseCase.Execute(context.Background(), updatepost.UpdatePostInput{ID: "post-id"})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if post != persisted {
				test.Errorf("Post = %+v, want the repository post %+v", post, persisted)
			}
		})

		test.Run("Should return the repository error and no post", func(test *testing.T) {
			repositoryError := errors.New("repository failed")
			updatePostUseCase := updatepost.NewUpdatePostUseCase(updatepost.NewUpdatePostUseCaseInput{
				PostRepository: &postRepositoryMock{
					UpdateFunc: func(ctx context.Context, input repositories.UpdatePostInput) (*entities.Post, error) {
						return nil, repositoryError
					},
				},
			})

			post, err := updatePostUseCase.Execute(context.Background(), updatepost.UpdatePostInput{ID: "post-id"})

			if !errors.Is(err, repositoryError) {
				test.Errorf("Error = %v, want %v", err, repositoryError)
			}
			if post != nil {
				test.Errorf("Post = %+v, want nil", post)
			}
		})
	})

}
