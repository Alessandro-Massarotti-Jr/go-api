package deletepostcomments

import (
	"context"
	"sync"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
)

// PostDeletedListener is the postDeleted listener ready to be bound to a topic.
// Its dependencies are wired on the first event, so importing this package
// (e.g. in tests) does not connect to the database.
var PostDeletedListener = &lazyPostDeletedListener{}

type lazyPostDeletedListener struct {
	once                                      sync.Once
	deletePostCommentsFromPostDeletedListener *DeletePostCommentsFromPostDeletedListener
	composeError                              error
}

func (lazyPostDeletedListener *lazyPostDeletedListener) Handle(ctx context.Context, event providers.Event) error {
	lazyPostDeletedListener.once.Do(func() {
		lazyPostDeletedListener.deletePostCommentsFromPostDeletedListener, lazyPostDeletedListener.composeError = composePostDeletedListener()
	})
	if lazyPostDeletedListener.composeError != nil {
		return lazyPostDeletedListener.composeError
	}

	return lazyPostDeletedListener.deletePostCommentsFromPostDeletedListener.Handle(ctx, event)
}

// composePostDeletedListener wires the dependencies: repository -> use case -> listener.
func composePostDeletedListener() (*DeletePostCommentsFromPostDeletedListener, error) {
	postCommentRepository, err := repositories.GetGormPostCommentRepositoryInstance()
	if err != nil {
		return nil, err
	}

	deletePostCommentsUseCase := NewDeletePostCommentsUseCase(NewDeletePostCommentsUseCaseInput{PostCommentRepository: postCommentRepository})

	return NewDeletePostCommentsFromPostDeletedListener(NewDeletePostCommentsFromPostDeletedListenerInput{
		DeletePostCommentsUseCase: deletePostCommentsUseCase,
	}), nil
}
