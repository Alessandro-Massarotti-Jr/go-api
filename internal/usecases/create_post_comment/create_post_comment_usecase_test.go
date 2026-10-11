package createpostcomment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/events"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	createpostcomment "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/create_post_comment"
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
	CreateFunc func(ctx context.Context, postComment *entities.PostComment) (*entities.PostComment, error)
	calls      int
}

var _ repositories.PostCommentRepository = (*postCommentRepositoryMock)(nil)

func (postCommentRepositoryMock *postCommentRepositoryMock) Create(ctx context.Context, postComment *entities.PostComment) (*entities.PostComment, error) {
	postCommentRepositoryMock.calls++
	if postCommentRepositoryMock.CreateFunc == nil {
		panic("postCommentRepositoryMock.Create called without CreateFunc")
	}
	return postCommentRepositoryMock.CreateFunc(ctx, postComment)
}

func (postCommentRepositoryMock *postCommentRepositoryMock) FindAllByPostId(ctx context.Context, input repositories.FindPostCommentsByPostIdInput) (*repositories.PaginatedPostComments, error) {
	panic("postCommentRepositoryMock.FindAllByPostId should not be called")
}

type eventProviderMock struct {
	PublishEventFunc func(ctx context.Context, input providers.PublishEventInput) error
	calls            int
}

var _ providers.EventProvider = (*eventProviderMock)(nil)

func (eventProviderMock *eventProviderMock) PublishEvent(ctx context.Context, input providers.PublishEventInput) error {
	eventProviderMock.calls++
	if eventProviderMock.PublishEventFunc == nil {
		panic("eventProviderMock.PublishEvent called without PublishEventFunc")
	}
	return eventProviderMock.PublishEventFunc(ctx, input)
}

func (eventProviderMock *eventProviderMock) AddListener(input providers.AddListenerInput) {
	panic("eventProviderMock.AddListener should not be called")
}

func (eventProviderMock *eventProviderMock) StartListening(ctx context.Context) error {
	panic("eventProviderMock.StartListening should not be called")
}

func existingPostRepository() *postRepositoryMock {
	return &postRepositoryMock{
		FindByIdFunc: func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
			return &entities.Post{ID: input.ID}, nil
		},
	}
}

func postCommentRepositoryReturningInput() *postCommentRepositoryMock {
	return &postCommentRepositoryMock{
		CreateFunc: func(ctx context.Context, postComment *entities.PostComment) (*entities.PostComment, error) {
			return postComment, nil
		},
	}
}

func eventProviderSucceeding() *eventProviderMock {
	return &eventProviderMock{
		PublishEventFunc: func(ctx context.Context, input providers.PublishEventInput) error { return nil },
	}
}

func TestCreatePostCommentUseCase(test *testing.T) {

	test.Run("Execute", func(test *testing.T) {
		test.Run("Should look up the post with the given ID", func(test *testing.T) {
			var received repositories.FindPostByIdInput
			createPostCommentUseCase := createpostcomment.NewCreatePostCommentUseCase(createpostcomment.NewCreatePostCommentUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindByIdFunc: func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
						received = input
						return &entities.Post{}, nil
					},
				},
				PostCommentRepository: postCommentRepositoryReturningInput(),
				EventProvider:         eventProviderSucceeding(),
			})

			createPostCommentUseCase.Execute(context.Background(), createpostcomment.CreatePostCommentInput{PostID: "post-id", Comment: "comment"})

			if received.ID != "post-id" {
				test.Errorf("ID = %q, want %q", received.ID, "post-id")
			}
		})

		test.Run("Should return the post lookup error without saving nor publishing", func(test *testing.T) {
			findError := errors.New("post not found")
			postCommentRepository := &postCommentRepositoryMock{}
			eventProvider := &eventProviderMock{}
			createPostCommentUseCase := createpostcomment.NewCreatePostCommentUseCase(createpostcomment.NewCreatePostCommentUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindByIdFunc: func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
						return nil, findError
					},
				},
				PostCommentRepository: postCommentRepository,
				EventProvider:         eventProvider,
			})

			postComment, err := createPostCommentUseCase.Execute(context.Background(), createpostcomment.CreatePostCommentInput{PostID: "post-id", Comment: "comment"})

			if !errors.Is(err, findError) {
				test.Errorf("Error = %v, want %v", err, findError)
			}
			if postComment != nil {
				test.Errorf("PostComment = %+v, want nil", postComment)
			}
			if postCommentRepository.calls != 0 || eventProvider.calls != 0 {
				test.Errorf("Expected no save nor publish, got %d saves and %d publishes", postCommentRepository.calls, eventProvider.calls)
			}
		})

		test.Run("Should save a comment with the given post ID and comment", func(test *testing.T) {
			var received *entities.PostComment
			createPostCommentUseCase := createpostcomment.NewCreatePostCommentUseCase(createpostcomment.NewCreatePostCommentUseCaseInput{
				PostRepository: existingPostRepository(),
				PostCommentRepository: &postCommentRepositoryMock{
					CreateFunc: func(ctx context.Context, postComment *entities.PostComment) (*entities.PostComment, error) {
						received = postComment
						return postComment, nil
					},
				},
				EventProvider: eventProviderSucceeding(),
			})

			_, err := createPostCommentUseCase.Execute(context.Background(), createpostcomment.CreatePostCommentInput{PostID: "post-id", Comment: "My comment"})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if received == nil {
				test.Fatal("Expected the repository Create to be called, it was not")
			}
			if received.PostID != "post-id" || received.Comment != "My comment" {
				test.Errorf("Saved comment = %+v, want PostID %q and Comment %q", received, "post-id", "My comment")
			}
		})

		test.Run("Should return the save error without publishing", func(test *testing.T) {
			repositoryError := errors.New("repository failed")
			eventProvider := &eventProviderMock{}
			createPostCommentUseCase := createpostcomment.NewCreatePostCommentUseCase(createpostcomment.NewCreatePostCommentUseCaseInput{
				PostRepository: existingPostRepository(),
				PostCommentRepository: &postCommentRepositoryMock{
					CreateFunc: func(ctx context.Context, postComment *entities.PostComment) (*entities.PostComment, error) {
						return nil, repositoryError
					},
				},
				EventProvider: eventProvider,
			})

			postComment, err := createPostCommentUseCase.Execute(context.Background(), createpostcomment.CreatePostCommentInput{PostID: "post-id", Comment: "comment"})

			if !errors.Is(err, repositoryError) {
				test.Errorf("Error = %v, want %v", err, repositoryError)
			}
			if postComment != nil {
				test.Errorf("PostComment = %+v, want nil", postComment)
			}
			if eventProvider.calls != 0 {
				test.Errorf("Expected no publish, got %d", eventProvider.calls)
			}
		})

		test.Run("Should publish postCommentCreated with the saved comment", func(test *testing.T) {
			persisted := &entities.PostComment{ID: "comment-id", PostID: "post-id", Comment: "persisted"}
			var received providers.PublishEventInput
			createPostCommentUseCase := createpostcomment.NewCreatePostCommentUseCase(createpostcomment.NewCreatePostCommentUseCaseInput{
				PostRepository: existingPostRepository(),
				PostCommentRepository: &postCommentRepositoryMock{
					CreateFunc: func(ctx context.Context, postComment *entities.PostComment) (*entities.PostComment, error) {
						return persisted, nil
					},
				},
				EventProvider: &eventProviderMock{
					PublishEventFunc: func(ctx context.Context, input providers.PublishEventInput) error {
						received = input
						return nil
					},
				},
			})

			createPostCommentUseCase.Execute(context.Background(), createpostcomment.CreatePostCommentInput{PostID: "post-id", Comment: "comment"})

			if received.Topic != events.PostCommentCreatedTopic {
				test.Errorf("Topic = %q, want %q", received.Topic, events.PostCommentCreatedTopic)
			}
			want := events.PostCommentCreatedEvent{ID: "comment-id", PostID: "post-id", Comment: "persisted", CreatedAt: persisted.CreatedAt}
			if received.Payload != want {
				test.Errorf("Payload = %+v, want %+v", received.Payload, want)
			}
		})

		test.Run("Should return the publish error and no comment", func(test *testing.T) {
			publishError := errors.New("publish failed")
			createPostCommentUseCase := createpostcomment.NewCreatePostCommentUseCase(createpostcomment.NewCreatePostCommentUseCaseInput{
				PostRepository:        existingPostRepository(),
				PostCommentRepository: postCommentRepositoryReturningInput(),
				EventProvider: &eventProviderMock{
					PublishEventFunc: func(ctx context.Context, input providers.PublishEventInput) error { return publishError },
				},
			})

			postComment, err := createPostCommentUseCase.Execute(context.Background(), createpostcomment.CreatePostCommentInput{PostID: "post-id", Comment: "comment"})

			if !errors.Is(err, publishError) {
				test.Errorf("Error = %v, want %v", err, publishError)
			}
			if postComment != nil {
				test.Errorf("PostComment = %+v, want nil", postComment)
			}
		})

		test.Run("Should return the comment returned by the repository", func(test *testing.T) {
			persisted := &entities.PostComment{ID: "comment-id", PostID: "post-id", Comment: "persisted"}
			createPostCommentUseCase := createpostcomment.NewCreatePostCommentUseCase(createpostcomment.NewCreatePostCommentUseCaseInput{
				PostRepository: existingPostRepository(),
				PostCommentRepository: &postCommentRepositoryMock{
					CreateFunc: func(ctx context.Context, postComment *entities.PostComment) (*entities.PostComment, error) {
						return persisted, nil
					},
				},
				EventProvider: eventProviderSucceeding(),
			})

			postComment, err := createPostCommentUseCase.Execute(context.Background(), createpostcomment.CreatePostCommentInput{PostID: "post-id", Comment: "comment"})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if postComment != persisted {
				test.Errorf("PostComment = %+v, want the repository comment %+v", postComment, persisted)
			}
		})

		test.Run("Should pass the received context to every dependency", func(test *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "request")

			var receivedContexts []context.Context
			createPostCommentUseCase := createpostcomment.NewCreatePostCommentUseCase(createpostcomment.NewCreatePostCommentUseCaseInput{
				PostRepository: &postRepositoryMock{
					FindByIdFunc: func(ctx context.Context, input repositories.FindPostByIdInput) (*entities.Post, error) {
						receivedContexts = append(receivedContexts, ctx)
						return &entities.Post{}, nil
					},
				},
				PostCommentRepository: &postCommentRepositoryMock{
					CreateFunc: func(ctx context.Context, postComment *entities.PostComment) (*entities.PostComment, error) {
						receivedContexts = append(receivedContexts, ctx)
						return postComment, nil
					},
				},
				EventProvider: &eventProviderMock{
					PublishEventFunc: func(ctx context.Context, input providers.PublishEventInput) error {
						receivedContexts = append(receivedContexts, ctx)
						return nil
					},
				},
			})

			createPostCommentUseCase.Execute(ctx, createpostcomment.CreatePostCommentInput{PostID: "post-id", Comment: "comment"})

			if len(receivedContexts) != 3 {
				test.Fatalf("Expected 3 dependency calls, got %d", len(receivedContexts))
			}
			for _, receivedCtx := range receivedContexts {
				if receivedCtx.Value(contextKey{}) != "request" {
					test.Errorf("Expected the dependency to receive the caller context, got %v", receivedCtx)
				}
			}
		})
	})

}
