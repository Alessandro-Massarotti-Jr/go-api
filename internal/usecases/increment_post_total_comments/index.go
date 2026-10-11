package incrementposttotalcomments

import (
	"context"
	"sync"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
)

// PostCommentCreatedListener is the postCommentCreated listener ready to be bound to a topic.
// Its dependencies are wired on the first event, so importing this package
// (e.g. in tests) does not connect to the database.
var PostCommentCreatedListener = &lazyPostCommentCreatedListener{}

type lazyPostCommentCreatedListener struct {
	once                                                     sync.Once
	incrementPostTotalCommentsFromPostCommentCreatedListener *IncrementPostTotalCommentsFromPostCommentCreatedListener
	composeError                                             error
}

func (lazyPostCommentCreatedListener *lazyPostCommentCreatedListener) Handle(ctx context.Context, event providers.Event) error {
	lazyPostCommentCreatedListener.once.Do(func() {
		lazyPostCommentCreatedListener.incrementPostTotalCommentsFromPostCommentCreatedListener, lazyPostCommentCreatedListener.composeError = composePostCommentCreatedListener()
	})
	if lazyPostCommentCreatedListener.composeError != nil {
		return lazyPostCommentCreatedListener.composeError
	}

	return lazyPostCommentCreatedListener.incrementPostTotalCommentsFromPostCommentCreatedListener.Handle(ctx, event)
}

// composePostCommentCreatedListener wires the dependencies: repository -> use case -> listener.
func composePostCommentCreatedListener() (*IncrementPostTotalCommentsFromPostCommentCreatedListener, error) {
	postRepository, err := repositories.GetMongoPostRepositoryInstance()
	if err != nil {
		return nil, err
	}

	incrementPostTotalCommentsUseCase := NewIncrementPostTotalCommentsUseCase(NewIncrementPostTotalCommentsUseCaseInput{PostRepository: postRepository})

	return NewIncrementPostTotalCommentsFromPostCommentCreatedListener(NewIncrementPostTotalCommentsFromPostCommentCreatedListenerInput{
		IncrementPostTotalCommentsUseCase: incrementPostTotalCommentsUseCase,
	}), nil
}
