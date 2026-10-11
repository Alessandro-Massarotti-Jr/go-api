package deletepost

import (
	"context"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/events"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
)

type DeletePostInput struct {
	ID entities.PostId
}

type DeletePostUseCase interface {
	Execute(ctx context.Context, input DeletePostInput) error
}

type deletePostUseCase struct {
	postRepository repositories.PostRepository
	eventProvider  providers.EventProvider
}

var _ DeletePostUseCase = (*deletePostUseCase)(nil)

type NewDeletePostUseCaseInput struct {
	PostRepository repositories.PostRepository
	EventProvider  providers.EventProvider
}

func NewDeletePostUseCase(input NewDeletePostUseCaseInput) DeletePostUseCase {
	return &deletePostUseCase{
		postRepository: input.PostRepository,
		eventProvider:  input.EventProvider,
	}
}

// Execute deletes the post and publishes postDeleted.
// The post comments are not touched here: they are removed by the postDeleted listener.
func (deletePostUseCase *deletePostUseCase) Execute(ctx context.Context, input DeletePostInput) error {
	if err := deletePostUseCase.postRepository.Delete(ctx, repositories.DeletePostInput{ID: input.ID}); err != nil {
		return err
	}

	return deletePostUseCase.eventProvider.PublishEvent(ctx, providers.PublishEventInput{
		Topic: events.PostDeletedTopic,
		Payload: events.PostDeletedEvent{
			ID:        string(input.ID),
			DeletedAt: time.Now().UTC(),
		},
	})
}
