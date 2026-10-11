package listpostcomments

import (
	"net/http"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/gin-gonic/gin"
)

type listPostCommentsUriRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type listPostCommentsQueryRequest struct {
	Cursor *string `form:"cursor"`
	Limit  *int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

type ListPostCommentsController struct {
	listPostCommentsUseCase ListPostCommentsUseCase
}

type NewListPostCommentsControllerInput struct {
	ListPostCommentsUseCase ListPostCommentsUseCase
}

func NewListPostCommentsController(input NewListPostCommentsControllerInput) *ListPostCommentsController {
	return &ListPostCommentsController{listPostCommentsUseCase: input.ListPostCommentsUseCase}
}

// Handle lists the comments of a post with cursor pagination, newest first.
//
//	@Summary	List post comments
//	@Tags		post comments
//	@Produce	json
//	@Param		id		path		string	true	"Post id (UUID)"	format(uuid)
//	@Param		cursor	query		string	false	"NextCursor returned by the previous page"
//	@Param		limit	query		int		false	"Page size"	minimum(1)	maximum(100)	default(20)
//	@Success	200		{object}	object{message=string,data=repositories.PaginatedPostComments}
//	@Failure	400		{object}	object{error=appErrors.AppError}
//	@Failure	404		{object}	object{error=appErrors.AppError}
//	@Failure	500		{object}	object{error=appErrors.AppError}
//	@Router		/posts/{id}/comments [get]
func (listPostCommentsController *ListPostCommentsController) Handle(context *gin.Context) {
	const location = "ListPostCommentsController.Handle"

	var uriRequest listPostCommentsUriRequest
	if err := context.ShouldBindUri(&uriRequest); err != nil {
		context.Error(appErrors.CreateInvalidRequestParamsError(appErrors.CreateInvalidRequestParamsErrorInput{
			Location:    location,
			OriginError: err,
		}))
		return
	}

	var queryRequest listPostCommentsQueryRequest
	if err := context.ShouldBindQuery(&queryRequest); err != nil {
		context.Error(appErrors.CreateInvalidRequestParamsError(appErrors.CreateInvalidRequestParamsErrorInput{
			Location:    location,
			OriginError: err,
		}))
		return
	}

	var limit int
	if queryRequest.Limit != nil {
		limit = *queryRequest.Limit
	}

	paginatedPostComments, err := listPostCommentsController.listPostCommentsUseCase.Execute(context.Request.Context(), ListPostCommentsInput{
		PostID: entities.PostId(uriRequest.ID),
		Cursor: queryRequest.Cursor,
		Limit:  limit,
	})
	if err != nil {
		context.Error(err)
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "Post comments listed successfully",
		"data":    paginatedPostComments,
	})
}
