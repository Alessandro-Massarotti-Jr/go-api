package incrementposttotalcomments_test

import (
	"context"
	"errors"
	"testing"

	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/events"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
	incrementposttotalcomments "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/increment_post_total_comments"
)

// incrementPostTotalCommentsUseCaseMock implements incrementposttotalcomments.IncrementPostTotalCommentsUseCase with a configurable Execute.
type incrementPostTotalCommentsUseCaseMock struct {
	ExecuteFunc func(ctx context.Context, input incrementposttotalcomments.IncrementPostTotalCommentsInput) error
	calls       int
}

func (incrementPostTotalCommentsUseCaseMock *incrementPostTotalCommentsUseCaseMock) Execute(ctx context.Context, input incrementposttotalcomments.IncrementPostTotalCommentsInput) error {
	incrementPostTotalCommentsUseCaseMock.calls++
	if incrementPostTotalCommentsUseCaseMock.ExecuteFunc == nil {
		panic("incrementPostTotalCommentsUseCaseMock.Execute called without ExecuteFunc")
	}
	return incrementPostTotalCommentsUseCaseMock.ExecuteFunc(ctx, input)
}

func newPostCommentCreatedEvent(payload string) providers.Event {
	return providers.Event{Topic: events.PostCommentCreatedTopic, Payload: []byte(payload)}
}

func TestIncrementPostTotalCommentsFromPostCommentCreatedListener(test *testing.T) {

	test.Run("Handle", func(test *testing.T) {
		test.Run("Should call the use case with the post ID from the event", func(test *testing.T) {
			var received incrementposttotalcomments.IncrementPostTotalCommentsInput
			incrementPostTotalCommentsUseCase := &incrementPostTotalCommentsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input incrementposttotalcomments.IncrementPostTotalCommentsInput) error {
					received = input
					return nil
				},
			}
			incrementPostTotalCommentsFromPostCommentCreatedListener := incrementposttotalcomments.NewIncrementPostTotalCommentsFromPostCommentCreatedListener(incrementposttotalcomments.NewIncrementPostTotalCommentsFromPostCommentCreatedListenerInput{
				IncrementPostTotalCommentsUseCase: incrementPostTotalCommentsUseCase,
			})

			err := incrementPostTotalCommentsFromPostCommentCreatedListener.Handle(context.Background(), newPostCommentCreatedEvent(`{"id":"comment-id","postId":"post-id","comment":"comment","createdAt":"2025-01-01T10:00:00Z"}`))

			if err != nil {
				test.Fatalf("Unexpected error: %v", err)
			}
			want := incrementposttotalcomments.IncrementPostTotalCommentsInput{PostID: "post-id"}
			if received != want {
				test.Errorf("Use case input = %+v, want %+v", received, want)
			}
		})

		test.Run("Should pass the received context to the use case", func(test *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "event")

			var receivedCtx context.Context
			incrementPostTotalCommentsUseCase := &incrementPostTotalCommentsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input incrementposttotalcomments.IncrementPostTotalCommentsInput) error {
					receivedCtx = ctx
					return nil
				},
			}
			incrementPostTotalCommentsFromPostCommentCreatedListener := incrementposttotalcomments.NewIncrementPostTotalCommentsFromPostCommentCreatedListener(incrementposttotalcomments.NewIncrementPostTotalCommentsFromPostCommentCreatedListenerInput{
				IncrementPostTotalCommentsUseCase: incrementPostTotalCommentsUseCase,
			})

			incrementPostTotalCommentsFromPostCommentCreatedListener.Handle(ctx, newPostCommentCreatedEvent(`{"postId":"post-id"}`))

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "event" {
				test.Errorf("Expected the use case to receive the listener context, got %v", receivedCtx)
			}
		})

		test.Run("Should return the use case error", func(test *testing.T) {
			useCaseError := errors.New("use case failed")
			incrementPostTotalCommentsFromPostCommentCreatedListener := incrementposttotalcomments.NewIncrementPostTotalCommentsFromPostCommentCreatedListener(incrementposttotalcomments.NewIncrementPostTotalCommentsFromPostCommentCreatedListenerInput{
				IncrementPostTotalCommentsUseCase: &incrementPostTotalCommentsUseCaseMock{
					ExecuteFunc: func(ctx context.Context, input incrementposttotalcomments.IncrementPostTotalCommentsInput) error {
						return useCaseError
					},
				},
			})

			err := incrementPostTotalCommentsFromPostCommentCreatedListener.Handle(context.Background(), newPostCommentCreatedEvent(`{"postId":"post-id"}`))

			if !errors.Is(err, useCaseError) {
				test.Errorf("Error = %v, want %v", err, useCaseError)
			}
		})

		invalidPayloads := []struct {
			name    string
			payload string
		}{
			{name: "payload is malformed JSON", payload: `{"postId":`},
			{name: "payload is empty", payload: ``},
			{name: "postId is missing", payload: `{"id":"comment-id"}`},
			{name: "postId is empty", payload: `{"postId":""}`},
			{name: "postId has wrong type", payload: `{"postId":1}`},
		}
		for _, tc := range invalidPayloads {
			test.Run("Should return an InvalidEventPayloadError and not call the use case when "+tc.name, func(test *testing.T) {
				incrementPostTotalCommentsUseCase := &incrementPostTotalCommentsUseCaseMock{}
				incrementPostTotalCommentsFromPostCommentCreatedListener := incrementposttotalcomments.NewIncrementPostTotalCommentsFromPostCommentCreatedListener(incrementposttotalcomments.NewIncrementPostTotalCommentsFromPostCommentCreatedListenerInput{
					IncrementPostTotalCommentsUseCase: incrementPostTotalCommentsUseCase,
				})

				err := incrementPostTotalCommentsFromPostCommentCreatedListener.Handle(context.Background(), newPostCommentCreatedEvent(tc.payload))

				if incrementPostTotalCommentsUseCase.calls != 0 {
					test.Errorf("Expected the use case not to be called, got %d calls", incrementPostTotalCommentsUseCase.calls)
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
