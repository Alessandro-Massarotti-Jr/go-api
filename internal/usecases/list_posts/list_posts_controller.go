package listposts

import (
	"net/http"

	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/gin-gonic/gin"
)

type listPostsRequest struct {
	Cursor *string `form:"cursor"`
	Limit  *int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

type ListPostsController struct {
	listPostsUseCase ListPostsUseCase
}

type NewListPostsControllerInput struct {
	ListPostsUseCase ListPostsUseCase
}

func NewListPostsController(input NewListPostsControllerInput) *ListPostsController {
	return &ListPostsController{listPostsUseCase: input.ListPostsUseCase}
}

func (listPostsController *ListPostsController) Handle(context *gin.Context) {
	const location = "ListPostsController.Handle"

	var request listPostsRequest
	if err := context.ShouldBindQuery(&request); err != nil {
		context.Error(appErrors.CreateInvalidRequestParamsError(appErrors.CreateInvalidRequestParamsErrorInput{
			Location:    location,
			OriginError: err,
		}))
		return
	}

	var limit int
	if request.Limit != nil {
		limit = *request.Limit
	}

	paginatedPosts, err := listPostsController.listPostsUseCase.Execute(context.Request.Context(), ListPostsInput{
		Cursor: request.Cursor,
		Limit:  limit,
	})
	if err != nil {
		context.Error(err)
		return
	}

	context.JSON(http.StatusOK, paginatedPosts)
}
