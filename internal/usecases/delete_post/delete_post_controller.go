package deletepost

import (
	"net/http"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/gin-gonic/gin"
)

type deletePostRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type DeletePostController struct {
	deletePostUseCase DeletePostUseCase
}

type NewDeletePostControllerInput struct {
	DeletePostUseCase DeletePostUseCase
}

func NewDeletePostController(input NewDeletePostControllerInput) *DeletePostController {
	return &DeletePostController{deletePostUseCase: input.DeletePostUseCase}
}

func (deletePostController *DeletePostController) Handle(context *gin.Context) {
	const location = "DeletePostController.Handle"

	var request deletePostRequest
	if err := context.ShouldBindUri(&request); err != nil {
		context.Error(appErrors.CreateInvalidRequestParamsError(appErrors.CreateInvalidRequestParamsErrorInput{
			Location:    location,
			OriginError: err,
		}))
		return
	}

	err := deletePostController.deletePostUseCase.Execute(context.Request.Context(), DeletePostInput{
		ID: entities.PostId(request.ID),
	})
	if err != nil {
		context.Error(err)
		return
	}

	context.Status(http.StatusNoContent)
}
