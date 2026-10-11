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

// Handle deletes a post by id.
//
//	@Summary	Delete post
//	@Tags		posts
//	@Produce	json
//	@Param		id	path		string	true	"Post id (UUID)"	format(uuid)
//	@Success	200	{object}	object{message=string,data=object}
//	@Failure	400	{object}	object{error=appErrors.AppError}
//	@Failure	404	{object}	object{error=appErrors.AppError}
//	@Failure	500	{object}	object{error=appErrors.AppError}
//	@Router		/posts/{id} [delete]
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

	context.JSON(http.StatusOK, gin.H{
		"message": "Post deleted successfully",
		"data":    nil,
	})
}
