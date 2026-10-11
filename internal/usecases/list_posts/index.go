package listposts

import (
	"sync"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	"github.com/gin-gonic/gin"
)

// Controller is the list posts controller ready to be bound to a route.
// Its dependencies are wired on the first request, so importing this package
// (e.g. in tests) does not connect to the database.
var Controller = &lazyListPostsController{}

type lazyListPostsController struct {
	once                sync.Once
	listPostsController *ListPostsController
	composeError        error
}

func (lazyListPostsController *lazyListPostsController) Handle(context *gin.Context) {
	lazyListPostsController.once.Do(func() {
		lazyListPostsController.listPostsController, lazyListPostsController.composeError = compose()
	})
	if lazyListPostsController.composeError != nil {
		context.Error(lazyListPostsController.composeError)
		return
	}

	lazyListPostsController.listPostsController.Handle(context)
}

// compose wires the list posts dependencies: repository -> use case -> controller.
func compose() (*ListPostsController, error) {
	postRepository, err := repositories.GetMongoPostRepositoryInstance()
	if err != nil {
		return nil, err
	}

	listPostsUseCase := NewListPostsUseCase(NewListPostsUseCaseInput{PostRepository: postRepository})

	return NewListPostsController(NewListPostsControllerInput{ListPostsUseCase: listPostsUseCase}), nil
}
