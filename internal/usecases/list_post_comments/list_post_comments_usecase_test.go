package listpostcomments_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	listpostcomments "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/list_post_comments"
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

func (postRepositoryMock *postRepositoryMock) IncrementTotalComments(ctx context.Context, input repositories.IncrementPostTotalCommentsInput) error {
	panic("postRepositoryMock.IncrementTotalComments should not be called")
}

type postCommentRepositoryMock struct {
	FindAllByPostIdFunc func(ctx context.Context, input repositories.FindPostCommentsByPostIdInput) (*repositories.PaginatedPostComments, error)
	calls               int
}

var _ repositories.PostCommentRepository = (*postCommentRepositoryMock)(nil)

func (postCommentRepositoryMock *postCommentRepositoryMock) FindAllByPostId(ctx context.Context, input repositories.FindPostCommentsByPostIdInput) (*repositories.PaginatedPostComments, error) {
	postCommentRepositoryMock.calls++
	if postCommentRepositoryMock.FindAllByPostIdFunc == nil {
		panic("postCommentRepositoryMock.FindAllByPostId called without FindAllByPostIdFunc")
	}
	return postCommentRepositoryMock.FindAllByPostIdFunc(ctx, input)
}

func (postCommentRepositoryMock *postCommentRepositoryMock) Create(ctx context.Context, postComment *entities.PostComment) (*entities.PostComment, error) {
	panic("postCommentRepositoryMock.Create should not be called")
}

func existingPostRepository() *postRepositoryMock {
	return &postRepositoryMock{
		FindByIdFunc: func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
			return &entities.Post{ID: input.ID}, nil
		},
	}
}

func TestListPostCommentsUseCase(test *testing.T) {

	test.Run("Execute", func(test *testing.T) {
		test.Run("Should look up the post with the given ID", func(test *testing.T) {
			var received repositories.FindPostByIdInput
			listPostCommentsUseCase := listpostcomments.NewListPostCommentsUseCase(listpostcomments.NewListPostCommentsUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindByIdFunc: func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
						received = input
						return &entities.Post{}, nil
					},
				},
				PostCommentRepository: &postCommentRepositoryMock{
					FindAllByPostIdFunc: func(ctx context.Context, input repositories.FindPostCommentsByPostIdInput) (*repositories.PaginatedPostComments, error) {
						return &repositories.PaginatedPostComments{}, nil
					},
				},
			})

			listPostCommentsUseCase.Execute(context.Background(), listpostcomments.ListPostCommentsInput{PostID: "post-id"})

			if received.ID != "post-id" {
				test.Errorf("ID = %q, want %q", received.ID, "post-id")
			}
		})

		test.Run("Should return the post lookup error without listing comments", func(test *testing.T) {
			findError := errors.New("post not found")
			postCommentRepository := &postCommentRepositoryMock{}
			listPostCommentsUseCase := listpostcomments.NewListPostCommentsUseCase(listpostcomments.NewListPostCommentsUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindByIdFunc: func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
						return nil, findError
					},
				},
				PostCommentRepository: postCommentRepository,
			})

			page, err := listPostCommentsUseCase.Execute(context.Background(), listpostcomments.ListPostCommentsInput{PostID: "post-id"})

			if !errors.Is(err, findError) {
				test.Errorf("Error = %v, want %v", err, findError)
			}
			if page != nil {
				test.Errorf("Page = %+v, want nil", page)
			}
			if postCommentRepository.calls != 0 {
				test.Errorf("Expected the comments not to be listed, got %d calls", postCommentRepository.calls)
			}
		})

		test.Run("Should send the post ID, cursor and limit to the repository", func(test *testing.T) {
			cursor := "cursor"
			var received repositories.FindPostCommentsByPostIdInput
			listPostCommentsUseCase := listpostcomments.NewListPostCommentsUseCase(listpostcomments.NewListPostCommentsUseCaseInput{
				PostRepository: existingPostRepository(),
				PostCommentRepository: &postCommentRepositoryMock{
					FindAllByPostIdFunc: func(ctx context.Context, input repositories.FindPostCommentsByPostIdInput) (*repositories.PaginatedPostComments, error) {
						received = input
						return &repositories.PaginatedPostComments{}, nil
					},
				},
			})

			listPostCommentsUseCase.Execute(context.Background(), listpostcomments.ListPostCommentsInput{PostID: "post-id", Cursor: &cursor, Limit: 10})

			want := repositories.FindPostCommentsByPostIdInput{PostID: "post-id", Cursor: &cursor, Limit: 10}
			if received != want {
				test.Errorf("Repository input = %+v, want %+v", received, want)
			}
		})

		test.Run("Should pass the received context to every dependency", func(test *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "request")

			var receivedContexts []context.Context
			listPostCommentsUseCase := listpostcomments.NewListPostCommentsUseCase(listpostcomments.NewListPostCommentsUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindByIdFunc: func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
						receivedContexts = append(receivedContexts, ctx)
						return &entities.Post{}, nil
					},
				},
				PostCommentRepository: &postCommentRepositoryMock{
					FindAllByPostIdFunc: func(ctx context.Context, input repositories.FindPostCommentsByPostIdInput) (*repositories.PaginatedPostComments, error) {
						receivedContexts = append(receivedContexts, ctx)
						return &repositories.PaginatedPostComments{}, nil
					},
				},
			})

			listPostCommentsUseCase.Execute(ctx, listpostcomments.ListPostCommentsInput{PostID: "post-id"})

			if len(receivedContexts) != 2 {
				test.Fatalf("Expected 2 dependency calls, got %d", len(receivedContexts))
			}
			for _, receivedCtx := range receivedContexts {
				if receivedCtx.Value(contextKey{}) != "request" {
					test.Errorf("Expected the dependency to receive the caller context, got %v", receivedCtx)
				}
			}
		})

		test.Run("Should return the page returned by the repository", func(test *testing.T) {
			persisted := &repositories.PaginatedPostComments{Items: []*entities.PostComment{{ID: "comment-id"}}}
			listPostCommentsUseCase := listpostcomments.NewListPostCommentsUseCase(listpostcomments.NewListPostCommentsUseCaseInput{
				PostRepository: existingPostRepository(),
				PostCommentRepository: &postCommentRepositoryMock{
					FindAllByPostIdFunc: func(ctx context.Context, input repositories.FindPostCommentsByPostIdInput) (*repositories.PaginatedPostComments, error) {
						return persisted, nil
					},
				},
			})

			page, err := listPostCommentsUseCase.Execute(context.Background(), listpostcomments.ListPostCommentsInput{PostID: "post-id"})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if page != persisted {
				test.Errorf("Page = %+v, want the repository page %+v", page, persisted)
			}
		})

		test.Run("Should return the repository error and no page", func(test *testing.T) {
			repositoryError := errors.New("repository failed")
			listPostCommentsUseCase := listpostcomments.NewListPostCommentsUseCase(listpostcomments.NewListPostCommentsUseCaseInput{
				PostRepository: existingPostRepository(),
				PostCommentRepository: &postCommentRepositoryMock{
					FindAllByPostIdFunc: func(ctx context.Context, input repositories.FindPostCommentsByPostIdInput) (*repositories.PaginatedPostComments, error) {
						return nil, repositoryError
					},
				},
			})

			page, err := listPostCommentsUseCase.Execute(context.Background(), listpostcomments.ListPostCommentsInput{PostID: "post-id"})

			if !errors.Is(err, repositoryError) {
				test.Errorf("Error = %v, want %v", err, repositoryError)
			}
			if page != nil {
				test.Errorf("Page = %+v, want nil", page)
			}
		})
	})

}
