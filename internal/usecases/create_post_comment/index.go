package createpostcomment

import (
	"sync"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	"github.com/gin-gonic/gin"
)

// Controller is the create post comment controller ready to be bound to a route.
// Its dependencies are wired on the first request, so importing this package
// (e.g. in tests) does not connect to the databases nor to RabbitMQ.
var Controller = &lazyCreatePostCommentController{}

type lazyCreatePostCommentController struct {
	once                        sync.Once
	createPostCommentController *CreatePostCommentController
	composeError                error
}

func (lazyCreatePostCommentController *lazyCreatePostCommentController) Handle(context *gin.Context) {
	lazyCreatePostCommentController.once.Do(func() {
		lazyCreatePostCommentController.createPostCommentController, lazyCreatePostCommentController.composeError = compose()
	})
	if lazyCreatePostCommentController.composeError != nil {
		context.Error(lazyCreatePostCommentController.composeError)
		return
	}

	lazyCreatePostCommentController.createPostCommentController.Handle(context)
}

// compose wires the create post comment dependencies: repositories + event provider -> use case -> controller.
func compose() (*CreatePostCommentController, error) {
	postRepository, err := repositories.GetMongoPostRepositoryInstance()
	if err != nil {
		return nil, err
	}

	postCommentRepository, err := repositories.GetGormPostCommentRepositoryInstance()
	if err != nil {
		return nil, err
	}

	eventProvider, err := providers.GetRabbitMQEventProviderInstance()
	if err != nil {
		return nil, err
	}

	createPostCommentUseCase := NewCreatePostCommentUseCase(NewCreatePostCommentUseCaseInput{
		PostRepository:        postRepository,
		PostCommentRepository: postCommentRepository,
		EventProvider:         eventProvider,
	})

	return NewCreatePostCommentController(NewCreatePostCommentControllerInput{CreatePostCommentUseCase: createPostCommentUseCase}), nil
}
