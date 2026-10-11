package incrementposttotalcomments

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/events"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
)

type IncrementPostTotalCommentsFromPostCommentCreatedListener struct {
	incrementPostTotalCommentsUseCase IncrementPostTotalCommentsUseCase
}

type NewIncrementPostTotalCommentsFromPostCommentCreatedListenerInput struct {
	IncrementPostTotalCommentsUseCase IncrementPostTotalCommentsUseCase
}

func NewIncrementPostTotalCommentsFromPostCommentCreatedListener(input NewIncrementPostTotalCommentsFromPostCommentCreatedListenerInput) *IncrementPostTotalCommentsFromPostCommentCreatedListener {
	return &IncrementPostTotalCommentsFromPostCommentCreatedListener{
		incrementPostTotalCommentsUseCase: input.IncrementPostTotalCommentsUseCase,
	}
}

// Handle increments the totalComments of the post referenced by a postCommentCreated event.
func (incrementPostTotalCommentsFromPostCommentCreatedListener *IncrementPostTotalCommentsFromPostCommentCreatedListener) Handle(ctx context.Context, event providers.Event) error {
	const location = "IncrementPostTotalCommentsFromPostCommentCreatedListener.Handle"

	var postCommentCreatedEvent events.PostCommentCreatedEvent
	if err := json.Unmarshal(event.Payload, &postCommentCreatedEvent); err != nil {
		return appErrors.CreateInvalidEventPayloadError(appErrors.CreateInvalidEventPayloadErrorInput{
			Location:    location,
			Topic:       event.Topic,
			OriginError: err,
		})
	}
	if postCommentCreatedEvent.PostID == "" {
		return appErrors.CreateInvalidEventPayloadError(appErrors.CreateInvalidEventPayloadErrorInput{
			Location:    location,
			Topic:       event.Topic,
			OriginError: errors.New("postId is required"),
		})
	}

	return incrementPostTotalCommentsFromPostCommentCreatedListener.incrementPostTotalCommentsUseCase.Execute(ctx, IncrementPostTotalCommentsInput{
		PostID: entities.PostId(postCommentCreatedEvent.PostID),
	})
}
