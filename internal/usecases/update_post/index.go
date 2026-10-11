package updatepost

import (
	"sync"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	"github.com/gin-gonic/gin"
)

// Controller is the update post controller ready to be bound to a route.
// Its dependencies are wired on the first request, so importing this package
// (e.g. in tests) does not connect to the database.
var Controller = &lazyUpdatePostController{}

type lazyUpdatePostController struct {
	once                 sync.Once
	updatePostController *UpdatePostController
	composeError         error
}

func (lazyUpdatePostController *lazyUpdatePostController) Handle(context *gin.Context) {
	lazyUpdatePostController.once.Do(func() {
		lazyUpdatePostController.updatePostController, lazyUpdatePostController.composeError = compose()
	})
	if lazyUpdatePostController.composeError != nil {
		context.Error(lazyUpdatePostController.composeError)
		return
	}

	lazyUpdatePostController.updatePostController.Handle(context)
}

// compose wires the update post dependencies: repository -> use case -> controller.
func compose() (*UpdatePostController, error) {
	postRepository, err := repositories.GetMongoPostRepositoryInstance()
	if err != nil {
		return nil, err
	}

	updatePostUseCase := NewUpdatePostUseCase(NewUpdatePostUseCaseInput{PostRepository: postRepository})

	return NewUpdatePostController(NewUpdatePostControllerInput{UpdatePostUseCase: updatePostUseCase}), nil
}
