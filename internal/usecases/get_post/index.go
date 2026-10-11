package getpost

import (
	"sync"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	"github.com/gin-gonic/gin"
)

// Controller is the get post controller ready to be bound to a route.
// Its dependencies are wired on the first request, so importing this package
// (e.g. in tests) does not connect to the database.
var Controller = &lazyGetPostController{}

type lazyGetPostController struct {
	once              sync.Once
	getPostController *GetPostController
	composeError      error
}

func (lazyGetPostController *lazyGetPostController) Handle(context *gin.Context) {
	lazyGetPostController.once.Do(func() {
		lazyGetPostController.getPostController, lazyGetPostController.composeError = compose()
	})
	if lazyGetPostController.composeError != nil {
		context.Error(lazyGetPostController.composeError)
		return
	}

	lazyGetPostController.getPostController.Handle(context)
}

// compose wires the get post dependencies: repository -> use case -> controller.
func compose() (*GetPostController, error) {
	postRepository, err := repositories.GetMongoPostRepositoryInstance()
	if err != nil {
		return nil, err
	}

	getPostUseCase := NewGetPostUseCase(NewGetPostUseCaseInput{PostRepository: postRepository})

	return NewGetPostController(NewGetPostControllerInput{GetPostUseCase: getPostUseCase}), nil
}
