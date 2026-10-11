package deletepost

import (
	"sync"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	"github.com/gin-gonic/gin"
)

// Controller is the delete post controller ready to be bound to a route.
// Its dependencies are wired on the first request, so importing this package
// (e.g. in tests) does not connect to the database nor to RabbitMQ.
var Controller = &lazyDeletePostController{}

type lazyDeletePostController struct {
	once                 sync.Once
	deletePostController *DeletePostController
	composeError         error
}

func (lazyDeletePostController *lazyDeletePostController) Handle(context *gin.Context) {
	lazyDeletePostController.once.Do(func() {
		lazyDeletePostController.deletePostController, lazyDeletePostController.composeError = compose()
	})
	if lazyDeletePostController.composeError != nil {
		context.Error(lazyDeletePostController.composeError)
		return
	}

	lazyDeletePostController.deletePostController.Handle(context)
}

// compose wires the delete post dependencies: repository + event provider -> use case -> controller.
func compose() (*DeletePostController, error) {
	postRepository, err := repositories.GetMongoPostRepositoryInstance()
	if err != nil {
		return nil, err
	}

	eventProvider, err := providers.GetRabbitMQEventProviderInstance()
	if err != nil {
		return nil, err
	}

	deletePostUseCase := NewDeletePostUseCase(NewDeletePostUseCaseInput{
		PostRepository: postRepository,
		EventProvider:  eventProvider,
	})

	return NewDeletePostController(NewDeletePostControllerInput{DeletePostUseCase: deletePostUseCase}), nil
}
