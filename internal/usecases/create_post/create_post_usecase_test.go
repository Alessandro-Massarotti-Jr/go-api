package createpost_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	createpost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/create_post"
	"github.com/google/uuid"
)

type postRepositoryMock struct {
	CreateFunc func(ctx context.Context, post *entities.Post) (*entities.Post, error)
}

var _ repositories.PostRepository = (*postRepositoryMock)(nil)

func (postRepositoryMock *postRepositoryMock) Create(ctx context.Context, post *entities.Post) (*entities.Post, error) {
	return postRepositoryMock.CreateFunc(ctx, post)
}

func (postRepositoryMock *postRepositoryMock) FindAll(ctx context.Context, input repositories.FindPostsInput) (*repositories.PaginatedPosts, error) {
	panic("postRepositoryMock.FindAll should not be called")
}

func (postRepositoryMock *postRepositoryMock) FindById(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
	panic("postRepositoryMock.FindById should not be called")
}

func (postRepositoryMock *postRepositoryMock) Update(ctx context.Context, input repositories.UpdatePostInput) (*entities.Post, error) {
	panic("postRepositoryMock.Update should not be called")
}

func (postRepositoryMock *postRepositoryMock) Delete(ctx context.Context, input repositories.DeletePostInput) error {
	panic("postRepositoryMock.Delete should not be called")
}

// postRepositoryReturningInput builds a repository mock whose Create returns the received post
// and stores it in received, so the test can inspect what the use case sent.
func postRepositoryReturningInput(received **entities.Post) *postRepositoryMock {
	return &postRepositoryMock{
		CreateFunc: func(ctx context.Context, post *entities.Post) (*entities.Post, error) {
			*received = post
			return post, nil
		},
	}
}

func TestCreatePostUseCase(test *testing.T) {

	test.Run("Execute", func(test *testing.T) {
		test.Run("Should send a post with the given title and content to the repository", func(test *testing.T) {
			var received *entities.Post
			createPostUseCase := createpost.NewCreatePostUseCase(createpost.NewCreatePostUseCaseInput{
				PostRepository: postRepositoryReturningInput(&received),
			})

			_, err := createPostUseCase.Execute(context.Background(), createpost.CreatePostInput{Title: "My title", Content: "My content"})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if received == nil {
				test.Fatal("Expected the repository Create to be called, it was not")
			}
			if received.Title != "My title" {
				test.Errorf("Title = %q, want %q", received.Title, "My title")
			}
			if received.Content != "My content" {
				test.Errorf("Content = %q, want %q", received.Content, "My content")
			}
		})

		test.Run("Should send a post with a generated UUID and dates set to now", func(test *testing.T) {
			var received *entities.Post
			createPostUseCase := createpost.NewCreatePostUseCase(createpost.NewCreatePostUseCaseInput{
				PostRepository: postRepositoryReturningInput(&received),
			})

			before := time.Now()
			_, err := createPostUseCase.Execute(context.Background(), createpost.CreatePostInput{Title: "title", Content: "content"})
			after := time.Now()

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if _, err := uuid.Parse(string(received.ID)); err != nil {
				test.Errorf("ID %q is not a valid UUID: %v", received.ID, err)
			}
			if received.CreatedAt.Before(before) || received.CreatedAt.After(after) {
				test.Errorf("CreatedAt = %v, want between %v and %v", received.CreatedAt, before, after)
			}
			if !received.UpdatedAt.Equal(received.CreatedAt) {
				test.Errorf("UpdatedAt = %v, want equal to CreatedAt %v", received.UpdatedAt, received.CreatedAt)
			}
		})

		test.Run("Should send a different ID on each call", func(test *testing.T) {
			var received *entities.Post
			createPostUseCase := createpost.NewCreatePostUseCase(createpost.NewCreatePostUseCaseInput{
				PostRepository: postRepositoryReturningInput(&received),
			})
			input := createpost.CreatePostInput{Title: "title", Content: "content"}

			createPostUseCase.Execute(context.Background(), input)
			firstID := received.ID
			createPostUseCase.Execute(context.Background(), input)

			if received.ID == firstID {
				test.Errorf("Expected different IDs, got %q twice", firstID)
			}
		})

		test.Run("Should pass the received context to the repository", func(test *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "request")

			var receivedCtx context.Context
			createPostUseCase := createpost.NewCreatePostUseCase(createpost.NewCreatePostUseCaseInput{
				PostRepository: &postRepositoryMock{
					CreateFunc: func(ctx context.Context, post *entities.Post) (*entities.Post, error) {
						receivedCtx = ctx
						return post, nil
					},
				},
			})

			createPostUseCase.Execute(ctx, createpost.CreatePostInput{Title: "title"})

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the repository to receive the caller context, got %v", receivedCtx)
			}
		})

		test.Run("Should return the post returned by the repository", func(test *testing.T) {
			persisted := &entities.Post{ID: "persisted-id", Title: "persisted"}
			createPostUseCase := createpost.NewCreatePostUseCase(createpost.NewCreatePostUseCaseInput{
				PostRepository: &postRepositoryMock{
					CreateFunc: func(ctx context.Context, post *entities.Post) (*entities.Post, error) {
						return persisted, nil
					},
				},
			})

			post, err := createPostUseCase.Execute(context.Background(), createpost.CreatePostInput{Title: "title"})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if post != persisted {
				test.Errorf("Post = %+v, want the repository post %+v", post, persisted)
			}
		})

		test.Run("Should return the repository error and no post", func(test *testing.T) {
			repositoryError := errors.New("repository failed")
			createPostUseCase := createpost.NewCreatePostUseCase(createpost.NewCreatePostUseCaseInput{
				PostRepository: &postRepositoryMock{
					CreateFunc: func(ctx context.Context, post *entities.Post) (*entities.Post, error) {
						return nil, repositoryError
					},
				},
			})

			post, err := createPostUseCase.Execute(context.Background(), createpost.CreatePostInput{Title: "title"})

			if !errors.Is(err, repositoryError) {
				test.Errorf("Error = %v, want %v", err, repositoryError)
			}
			if post != nil {
				test.Errorf("Post = %+v, want nil", post)
			}
		})
	})

}
