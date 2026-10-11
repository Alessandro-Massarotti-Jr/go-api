package createpost

import (
	"sync"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	"github.com/gin-gonic/gin"
)

// Controller is the create post controller ready to be bound to a route.
// Its dependencies are wired on the first request, so importing this package
// (e.g. in tests) does not connect to the database.
var Controller = &lazyCreatePostController{}

type lazyCreatePostController struct {
	once                 sync.Once
	createPostController *CreatePostController
	composeError         error
}

func (lazyCreatePostController *lazyCreatePostController) Handle(context *gin.Context) {
	lazyCreatePostController.once.Do(func() {
		lazyCreatePostController.createPostController, lazyCreatePostController.composeError = compose()
	})
	if lazyCreatePostController.composeError != nil {
		context.Error(lazyCreatePostController.composeError)
		return
	}

	lazyCreatePostController.createPostController.Handle(context)
}

// compose wires the create post dependencies: repository -> use case -> controller.
func compose() (*CreatePostController, error) {
	postRepository, err := repositories.GetMongoPostRepositoryInstance()
	if err != nil {
		return nil, err
	}

	createPostUseCase := NewCreatePostUseCase(NewCreatePostUseCaseInput{PostRepository: postRepository})

	return NewCreatePostController(NewCreatePostControllerInput{CreatePostUseCase: createPostUseCase}), nil
}
