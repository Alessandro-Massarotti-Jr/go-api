package deletepostcomments_test

import (
	"context"
	"errors"
	"testing"

	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/events"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
	deletepostcomments "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/delete_post_comments"
)

// deletePostCommentsUseCaseMock implements deletepostcomments.DeletePostCommentsUseCase with a configurable Execute.
type deletePostCommentsUseCaseMock struct {
	ExecuteFunc func(ctx context.Context, input deletepostcomments.DeletePostCommentsInput) error
	calls       int
}

func (deletePostCommentsUseCaseMock *deletePostCommentsUseCaseMock) Execute(ctx context.Context, input deletepostcomments.DeletePostCommentsInput) error {
	deletePostCommentsUseCaseMock.calls++
	if deletePostCommentsUseCaseMock.ExecuteFunc == nil {
		panic("deletePostCommentsUseCaseMock.Execute called without ExecuteFunc")
	}
	return deletePostCommentsUseCaseMock.ExecuteFunc(ctx, input)
}

func newPostDeletedEvent(payload string) providers.Event {
	return providers.Event{Topic: events.PostDeletedTopic, Payload: []byte(payload)}
}

func TestDeletePostCommentsFromPostDeletedListener(test *testing.T) {

	test.Run("Handle", func(test *testing.T) {
		test.Run("Should call the use case with the post ID from the event", func(test *testing.T) {
			var received deletepostcomments.DeletePostCommentsInput
			deletePostCommentsUseCase := &deletePostCommentsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input deletepostcomments.DeletePostCommentsInput) error {
					received = input
					return nil
				},
			}
			deletePostCommentsFromPostDeletedListener := deletepostcomments.NewDeletePostCommentsFromPostDeletedListener(deletepostcomments.NewDeletePostCommentsFromPostDeletedListenerInput{
				DeletePostCommentsUseCase: deletePostCommentsUseCase,
			})

			err := deletePostCommentsFromPostDeletedListener.Handle(context.Background(), newPostDeletedEvent(`{"id":"post-id","deletedAt":"2025-01-01T10:00:00Z"}`))

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			want := deletepostcomments.DeletePostCommentsInput{PostID: "post-id"}
			if received != want {
				test.Errorf("Use case input = %+v, want %+v", received, want)
			}
		})

		test.Run("Should pass the received context to the use case", func(test *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "event")

			var receivedCtx context.Context
			deletePostCommentsUseCase := &deletePostCommentsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input deletepostcomments.DeletePostCommentsInput) error {
					receivedCtx = ctx
					return nil
				},
			}
			deletePostCommentsFromPostDeletedListener := deletepostcomments.NewDeletePostCommentsFromPostDeletedListener(deletepostcomments.NewDeletePostCommentsFromPostDeletedListenerInput{
				DeletePostCommentsUseCase: deletePostCommentsUseCase,
			})

			deletePostCommentsFromPostDeletedListener.Handle(ctx, newPostDeletedEvent(`{"id":"post-id"}`))

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "event" {
				test.Errorf("Expected the use case to receive the listener context, got %v", receivedCtx)
			}
		})

		test.Run("Should return the use case error", func(test *testing.T) {
			useCaseError := errors.New("use case failed")
			deletePostCommentsFromPostDeletedListener := deletepostcomments.NewDeletePostCommentsFromPostDeletedListener(deletepostcomments.NewDeletePostCommentsFromPostDeletedListenerInput{
				DeletePostCommentsUseCase: &deletePostCommentsUseCaseMock{
					ExecuteFunc: func(ctx context.Context, input deletepostcomments.DeletePostCommentsInput) error {
						return useCaseError
					},
				},
			})

			err := deletePostCommentsFromPostDeletedListener.Handle(context.Background(), newPostDeletedEvent(`{"id":"post-id"}`))

			if !errors.Is(err, useCaseError) {
				test.Errorf("Error = %v, want %v", err, useCaseError)
			}
		})

		invalidPayloads := []struct {
			name    string
			payload string
		}{
			{name: "payload is malformed JSON", payload: `{"id":`},
			{name: "payload is empty", payload: ``},
			{name: "id is missing", payload: `{"deletedAt":"2025-01-01T10:00:00Z"}`},
			{name: "id is empty", payload: `{"id":""}`},
			{name: "id has wrong type", payload: `{"id":1}`},
		}
		for _, tc := range invalidPayloads {
			test.Run("Should return an InvalidEventPayloadError and not call the use case when "+tc.name, func(test *testing.T) {
				deletePostCommentsUseCase := &deletePostCommentsUseCaseMock{}
				deletePostCommentsFromPostDeletedListener := deletepostcomments.NewDeletePostCommentsFromPostDeletedListener(deletepostcomments.NewDeletePostCommentsFromPostDeletedListenerInput{
					DeletePostCommentsUseCase: deletePostCommentsUseCase,
				})

				err := deletePostCommentsFromPostDeletedListener.Handle(context.Background(), newPostDeletedEvent(tc.payload))

				if deletePostCommentsUseCase.calls != 0 {
					test.Errorf("Expected the use case not to be called, got %d calls", deletePostCommentsUseCase.calls)
				}
				appError, isAppError := errors.AsType[*appErrors.AppError](err)
				if !isAppError {
					test.Fatalf("Error = %v, want an AppError", err)
				}
				if appError.Name != "InvalidEventPayloadError" {
					test.Errorf("Name = %q, want %q", appError.Name, "InvalidEventPayloadError")
				}
			})
		}
	})

}
