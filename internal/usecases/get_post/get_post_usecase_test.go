package getpost_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	getpost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/get_post"
)

type postRepositoryMock struct {
	FindByIdFunc func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error)
}

var _ repositories.PostRepository = (*postRepositoryMock)(nil)

func (postRepositoryMock *postRepositoryMock) FindById(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
	return postRepositoryMock.FindByIdFunc(ctx, input)
}

func (postRepositoryMock *postRepositoryMock) FindAll(ctx context.Context, input repositories.FindPostsInput) (*repositories.PaginatedPosts, error) {
	panic("postRepositoryMock.FindAll should not be called")
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

func TestGetPostUseCase(test *testing.T) {

	test.Run("Execute", func(test *testing.T) {
		test.Run("Should send the given ID to the repository", func(test *testing.T) {
			var received repositories.FindPostByIdInput
			getPostUseCase := getpost.NewGetPostUseCase(getpost.NewGetPostUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindByIdFunc: func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
						received = input
						return &entities.Post{}, nil
					},
				},
			})

			_, err := getPostUseCase.Execute(context.Background(), getpost.GetPostInput{ID: "post-id"})

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
			getPostUseCase := getpost.NewGetPostUseCase(getpost.NewGetPostUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindByIdFunc: func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
						receivedCtx = ctx
						return &entities.Post{}, nil
					},
				},
			})

			getPostUseCase.Execute(ctx, getpost.GetPostInput{ID: "post-id"})

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the repository to receive the caller context, got %v", receivedCtx)
			}
		})

		test.Run("Should return the post returned by the repository", func(test *testing.T) {
			persisted := &entities.Post{ID: "post-id", Title: "persisted"}
			getPostUseCase := getpost.NewGetPostUseCase(getpost.NewGetPostUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindByIdFunc: func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
						return persisted, nil
					},
				},
			})

			post, err := getPostUseCase.Execute(context.Background(), getpost.GetPostInput{ID: "post-id"})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if post != persisted {
				test.Errorf("Post = %+v, want the repository post %+v", post, persisted)
			}
		})

		test.Run("Should return the repository error and no post", func(test *testing.T) {
			repositoryError := errors.New("repository failed")
			getPostUseCase := getpost.NewGetPostUseCase(getpost.NewGetPostUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindByIdFunc: func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
						return nil, repositoryError
					},
				},
			})

			post, err := getPostUseCase.Execute(context.Background(), getpost.GetPostInput{ID: "post-id"})

			if !errors.Is(err, repositoryError) {
				test.Errorf("Error = %v, want %v", err, repositoryError)
			}
			if post != nil {
				test.Errorf("Post = %+v, want nil", post)
			}
		})
	})

}
