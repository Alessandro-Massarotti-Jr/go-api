package listposts_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	listposts "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/list_posts"
)

type postRepositoryMock struct {
	FindAllFunc func(ctx context.Context, input repositories.FindPostsInput) (*repositories.PaginatedPosts, error)
}

var _ repositories.PostRepository = (*postRepositoryMock)(nil)

func (postRepositoryMock *postRepositoryMock) FindAll(ctx context.Context, input repositories.FindPostsInput) (*repositories.PaginatedPosts, error) {
	return postRepositoryMock.FindAllFunc(ctx, input)
}

func (postRepositoryMock *postRepositoryMock) Create(ctx context.Context, post *entities.Post) (*entities.Post, error) {
	panic("postRepositoryMock.Create should not be called")
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

func (postRepositoryMock *postRepositoryMock) IncrementTotalComments(ctx context.Context, input repositories.IncrementPostTotalCommentsInput) error {
	panic("postRepositoryMock.IncrementTotalComments should not be called")
}

func TestListPostsUseCase(test *testing.T) {

	test.Run("Execute", func(test *testing.T) {
		test.Run("Should send the cursor and limit to the repository", func(test *testing.T) {
			var received repositories.FindPostsInput
			listPostsUseCase := listposts.NewListPostsUseCase(listposts.NewListPostsUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindAllFunc: func(ctx context.Context, input repositories.FindPostsInput) (*repositories.PaginatedPosts, error) {
						received = input
						return &repositories.PaginatedPosts{}, nil
					},
				},
			})
			cursor := "next-cursor"

			_, err := listPostsUseCase.Execute(context.Background(), listposts.ListPostsInput{Cursor: &cursor, Limit: 10})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if received.Cursor == nil || *received.Cursor != cursor {
				test.Errorf("Cursor = %v, want %q", received.Cursor, cursor)
			}
			if received.Limit != 10 {
				test.Errorf("Limit = %d, want %d", received.Limit, 10)
			}
		})

		test.Run("Should send a nil cursor and zero limit when not given", func(test *testing.T) {
			var received repositories.FindPostsInput
			listPostsUseCase := listposts.NewListPostsUseCase(listposts.NewListPostsUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindAllFunc: func(ctx context.Context, input repositories.FindPostsInput) (*repositories.PaginatedPosts, error) {
						received = input
						return &repositories.PaginatedPosts{}, nil
					},
				},
			})

			listPostsUseCase.Execute(context.Background(), listposts.ListPostsInput{})

			if received.Cursor != nil {
				test.Errorf("Cursor = %v, want nil", received.Cursor)
			}
			if received.Limit != 0 {
				test.Errorf("Limit = %d, want %d", received.Limit, 0)
			}
		})

		test.Run("Should pass the received context to the repository", func(test *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "request")

			var receivedCtx context.Context
			listPostsUseCase := listposts.NewListPostsUseCase(listposts.NewListPostsUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindAllFunc: func(ctx context.Context, input repositories.FindPostsInput) (*repositories.PaginatedPosts, error) {
						receivedCtx = ctx
						return &repositories.PaginatedPosts{}, nil
					},
				},
			})

			listPostsUseCase.Execute(ctx, listposts.ListPostsInput{})

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the repository to receive the caller context, got %v", receivedCtx)
			}
		})

		test.Run("Should return the page returned by the repository", func(test *testing.T) {
			nextCursor := "next-cursor"
			persisted := &repositories.PaginatedPosts{
				Items:      []*entities.Post{{ID: "post-id", Title: "title"}},
				NextCursor: &nextCursor,
			}
			listPostsUseCase := listposts.NewListPostsUseCase(listposts.NewListPostsUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindAllFunc: func(ctx context.Context, input repositories.FindPostsInput) (*repositories.PaginatedPosts, error) {
						return persisted, nil
					},
				},
			})

			paginatedPosts, err := listPostsUseCase.Execute(context.Background(), listposts.ListPostsInput{})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if paginatedPosts != persisted {
				test.Errorf("PaginatedPosts = %+v, want the repository page %+v", paginatedPosts, persisted)
			}
		})

		test.Run("Should return the repository error and no page", func(test *testing.T) {
			repositoryError := errors.New("repository failed")
			listPostsUseCase := listposts.NewListPostsUseCase(listposts.NewListPostsUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindAllFunc: func(ctx context.Context, input repositories.FindPostsInput) (*repositories.PaginatedPosts, error) {
						return nil, repositoryError
					},
				},
			})

			paginatedPosts, err := listPostsUseCase.Execute(context.Background(), listposts.ListPostsInput{})

			if !errors.Is(err, repositoryError) {
				test.Errorf("Error = %v, want %v", err, repositoryError)
			}
			if paginatedPosts != nil {
				test.Errorf("PaginatedPosts = %+v, want nil", paginatedPosts)
			}
		})
	})

}
