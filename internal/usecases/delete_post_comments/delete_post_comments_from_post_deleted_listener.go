package deletepostcomments

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/events"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
)

type DeletePostCommentsFromPostDeletedListener struct {
	deletePostCommentsUseCase DeletePostCommentsUseCase
}

type NewDeletePostCommentsFromPostDeletedListenerInput struct {
	DeletePostCommentsUseCase DeletePostCommentsUseCase
}

func NewDeletePostCommentsFromPostDeletedListener(input NewDeletePostCommentsFromPostDeletedListenerInput) *DeletePostCommentsFromPostDeletedListener {
	return &DeletePostCommentsFromPostDeletedListener{
		deletePostCommentsUseCase: input.DeletePostCommentsUseCase,
	}
}

// Handle deletes every comment of the post referenced by a postDeleted event.
func (deletePostCommentsFromPostDeletedListener *DeletePostCommentsFromPostDeletedListener) Handle(ctx context.Context, event providers.Event) error {
	const location = "DeletePostCommentsFromPostDeletedListener.Handle"

	var postDeletedEvent events.PostDeletedEvent
	if err := json.Unmarshal(event.Payload, &postDeletedEvent); err != nil {
		return appErrors.CreateInvalidEventPayloadError(appErrors.CreateInvalidEventPayloadErrorInput{
			Location:    location,
			Topic:       event.Topic,
			OriginError: err,
		})
	}
	if postDeletedEvent.ID == "" {
		return appErrors.CreateInvalidEventPayloadError(appErrors.CreateInvalidEventPayloadErrorInput{
			Location:    location,
			Topic:       event.Topic,
			OriginError: errors.New("id is required"),
		})
	}

	return deletePostCommentsFromPostDeletedListener.deletePostCommentsUseCase.Execute(ctx, DeletePostCommentsInput{
		PostID: entities.PostId(postDeletedEvent.ID),
	})
}
