package listpostcomments

import (
	"sync"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	"github.com/gin-gonic/gin"
)

// Controller is the list post comments controller ready to be bound to a route.
// Its dependencies are wired on the first request, so importing this package
// (e.g. in tests) does not connect to the databases.
var Controller = &lazyListPostCommentsController{}

type lazyListPostCommentsController struct {
	once                       sync.Once
	listPostCommentsController *ListPostCommentsController
	composeError               error
}

func (lazyListPostCommentsController *lazyListPostCommentsController) Handle(context *gin.Context) {
	lazyListPostCommentsController.once.Do(func() {
		lazyListPostCommentsController.listPostCommentsController, lazyListPostCommentsController.composeError = compose()
	})
	if lazyListPostCommentsController.composeError != nil {
		context.Error(lazyListPostCommentsController.composeError)
		return
	}

	lazyListPostCommentsController.listPostCommentsController.Handle(context)
}

// compose wires the list post comments dependencies: repositories -> use case -> controller.
func compose() (*ListPostCommentsController, error) {
	postRepository, err := repositories.GetMongoPostRepositoryInstance()
	if err != nil {
		return nil, err
	}

	postCommentRepository, err := repositories.GetGormPostCommentRepositoryInstance()
	if err != nil {
		return nil, err
	}

	listPostCommentsUseCase := NewListPostCommentsUseCase(NewListPostCommentsUseCaseInput{
		PostRepository:        postRepository,
		PostCommentRepository: postCommentRepository,
	})

	return NewListPostCommentsController(NewListPostCommentsControllerInput{ListPostCommentsUseCase: listPostCommentsUseCase}), nil
}
