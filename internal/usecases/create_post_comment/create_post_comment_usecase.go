package createpostcomment

import (
	"context"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/events"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
)

type CreatePostCommentInput struct {
	PostID  entities.PostId
	Comment string
}

type CreatePostCommentUseCase interface {
	Execute(ctx context.Context, input CreatePostCommentInput) (*entities.PostComment, error)
}

type createPostCommentUseCase struct {
	postRepository        repositories.PostRepository
	postCommentRepository repositories.PostCommentRepository
	eventProvider         providers.EventProvider
}

var _ CreatePostCommentUseCase = (*createPostCommentUseCase)(nil)

type NewCreatePostCommentUseCaseInput struct {
	PostRepository        repositories.PostRepository
	PostCommentRepository repositories.PostCommentRepository
	EventProvider         providers.EventProvider
}

func NewCreatePostCommentUseCase(input NewCreatePostCommentUseCaseInput) CreatePostCommentUseCase {
	return &createPostCommentUseCase{
		postRepository:        input.PostRepository,
		postCommentRepository: input.PostCommentRepository,
		eventProvider:         input.EventProvider,
	}
}

// Execute saves the comment and publishes postCommentCreated.
// The post totalComments is not touched here: it is updated by the postCommentCreated listener.
func (createPostCommentUseCase *createPostCommentUseCase) Execute(ctx context.Context, input CreatePostCommentInput) (*entities.PostComment, error) {
	if _, err := createPostCommentUseCase.postRepository.FindById(ctx, repositories.FindPostByIdInput{ID: input.PostID}); err != nil {
		return nil, err
	}

	postComment := entities.CreatePostComment(entities.CreatePostCommentInput{
		PostID:  input.PostID,
		Comment: input.Comment,
	})

	createdPostComment, err := createPostCommentUseCase.postCommentRepository.Create(ctx, postComment)
	if err != nil {
		return nil, err
	}

	err = createPostCommentUseCase.eventProvider.PublishEvent(ctx, providers.PublishEventInput{
		Topic: events.PostCommentCreatedTopic,
		Payload: events.PostCommentCreatedEvent{
			ID:        string(createdPostComment.ID),
			PostID:    string(createdPostComment.PostID),
			Comment:   createdPostComment.Comment,
			CreatedAt: createdPostComment.CreatedAt,
		},
	})
	if err != nil {
		return nil, err
	}

	return createdPostComment, nil
}
