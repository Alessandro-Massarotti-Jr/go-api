package deletepost_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/events"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
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

func eventProviderSucceeding() *eventProviderMock {
	return &eventProviderMock{
		PublishEventFunc: func(ctx context.Context, input providers.PublishEventInput) error { return nil },
	}
}

func postRepositorySucceeding() *postRepositoryMock {
	return &postRepositoryMock{
		DeleteFunc: func(ctx context.Context, input repositories.DeletePostInput) error { return nil },
	}
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
				EventProvider: eventProviderSucceeding(),
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
				EventProvider: eventProviderSucceeding(),
			})

			deletePostUseCase.Execute(ctx, deletepost.DeletePostInput{ID: "post-id"})

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the repository to receive the caller context, got %v", receivedCtx)
			}
		})

		test.Run("Should return the repository error and not publish the event", func(test *testing.T) {
			repositoryError := errors.New("repository failed")
			eventProvider := &eventProviderMock{}
			deletePostUseCase := deletepost.NewDeletePostUseCase(deletepost.NewDeletePostUseCaseInput{
				PostRepository: &postRepositoryMock{
					DeleteFunc: func(ctx context.Context, input repositories.DeletePostInput) error {
						return repositoryError
					},
				},
				EventProvider: eventProvider,
			})

			err := deletePostUseCase.Execute(context.Background(), deletepost.DeletePostInput{ID: "post-id"})

			if !errors.Is(err, repositoryError) {
				test.Errorf("Error = %v, want %v", err, repositoryError)
			}
			if eventProvider.calls != 0 {
				test.Errorf("Expected no publish, got %d", eventProvider.calls)
			}
		})

		test.Run("Should publish postDeleted with the deleted post ID", func(test *testing.T) {
			var received providers.PublishEventInput
			before := time.Now().UTC()
			deletePostUseCase := deletepost.NewDeletePostUseCase(deletepost.NewDeletePostUseCaseInput{
				PostRepository: postRepositorySucceeding(),
				EventProvider: &eventProviderMock{
					PublishEventFunc: func(ctx context.Context, input providers.PublishEventInput) error {
						received = input
						return nil
					},
				},
			})

			err := deletePostUseCase.Execute(context.Background(), deletepost.DeletePostInput{ID: "post-id"})

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			if received.Topic != events.PostDeletedTopic {
				test.Errorf("Topic = %q, want %q", received.Topic, events.PostDeletedTopic)
			}
			payload, isPostDeletedEvent := received.Payload.(events.PostDeletedEvent)
			if !isPostDeletedEvent {
				test.Fatalf("Payload = %T, want events.PostDeletedEvent", received.Payload)
			}
			if payload.ID != "post-id" {
				test.Errorf("ID = %q, want %q", payload.ID, "post-id")
			}
			if payload.DeletedAt.Before(before) || payload.DeletedAt.After(time.Now().UTC()) {
				test.Errorf("DeletedAt = %v, want a time between %v and now", payload.DeletedAt, before)
			}
		})

		test.Run("Should pass the received context to the event provider", func(test *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "request")

			var receivedCtx context.Context
			deletePostUseCase := deletepost.NewDeletePostUseCase(deletepost.NewDeletePostUseCaseInput{
				PostRepository: postRepositorySucceeding(),
				EventProvider: &eventProviderMock{
					PublishEventFunc: func(ctx context.Context, input providers.PublishEventInput) error {
						receivedCtx = ctx
						return nil
					},
				},
			})

			deletePostUseCase.Execute(ctx, deletepost.DeletePostInput{ID: "post-id"})

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the event provider to receive the caller context, got %v", receivedCtx)
			}
		})

		test.Run("Should return the event provider error", func(test *testing.T) {
			publishError := errors.New("publish failed")
			deletePostUseCase := deletepost.NewDeletePostUseCase(deletepost.NewDeletePostUseCaseInput{
				PostRepository: postRepositorySucceeding(),
				EventProvider: &eventProviderMock{
					PublishEventFunc: func(ctx context.Context, input providers.PublishEventInput) error { return publishError },
				},
			})

			err := deletePostUseCase.Execute(context.Background(), deletepost.DeletePostInput{ID: "post-id"})

			if !errors.Is(err, publishError) {
				test.Errorf("Error = %v, want %v", err, publishError)
			}
		})
	})

}
