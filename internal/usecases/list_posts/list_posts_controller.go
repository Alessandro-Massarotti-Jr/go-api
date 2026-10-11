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

// Handle lists posts with cursor pagination.
//
//	@Summary	List posts
//	@Tags		posts
//	@Produce	json
//	@Param		cursor	query		string	false	"NextCursor returned by the previous page"
//	@Param		limit	query		int		false	"Page size"	minimum(1)	maximum(100)	default(20)
//	@Success	200		{object}	repositories.PaginatedPosts
//	@Failure	400		{object}	appErrors.AppError
//	@Failure	500		{object}	appErrors.AppError
//	@Router		/posts [get]
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
