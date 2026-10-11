package updatepost

import (
	"errors"
	"net/http"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/gin-gonic/gin"
)

type updatePostUriRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type updatePostBodyRequest struct {
	Title   *string `json:"title" binding:"omitempty,min=1"`
	Content *string `json:"content" binding:"omitempty,min=1"`
}

type UpdatePostController struct {
	updatePostUseCase UpdatePostUseCase
}

type NewUpdatePostControllerInput struct {
	UpdatePostUseCase UpdatePostUseCase
}

func NewUpdatePostController(input NewUpdatePostControllerInput) *UpdatePostController {
	return &UpdatePostController{updatePostUseCase: input.UpdatePostUseCase}
}

// Handle partially updates a post. At least one field must be provided.
//
//	@Summary	Update post
//	@Tags		posts
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string					true	"Post id (UUID)"	format(uuid)
//	@Param		body	body		updatePostBodyRequest	true	"Fields to update"
//	@Success	200		{object}	object{message=string,data=entities.Post}
//	@Failure	400		{object}	object{error=appErrors.AppError}
//	@Failure	404		{object}	object{error=appErrors.AppError}
//	@Failure	500		{object}	object{error=appErrors.AppError}
//	@Router		/posts/{id} [patch]
func (updatePostController *UpdatePostController) Handle(context *gin.Context) {
	const location = "UpdatePostController.Handle"

	var uriRequest updatePostUriRequest
	if err := context.ShouldBindUri(&uriRequest); err != nil {
		context.Error(appErrors.CreateInvalidRequestParamsError(appErrors.CreateInvalidRequestParamsErrorInput{
			Location:    location,
			OriginError: err,
		}))
		return
	}

	var bodyRequest updatePostBodyRequest
	if err := context.ShouldBindJSON(&bodyRequest); err != nil {
		context.Error(appErrors.CreateInvalidRequestParamsError(appErrors.CreateInvalidRequestParamsErrorInput{
			Location:    location,
			OriginError: err,
		}))
		return
	}
	if bodyRequest.Title == nil && bodyRequest.Content == nil {
		context.Error(appErrors.CreateInvalidRequestParamsError(appErrors.CreateInvalidRequestParamsErrorInput{
			Location:    location,
			OriginError: errors.New("at least one of title or content must be provided"),
		}))
		return
	}

	post, err := updatePostController.updatePostUseCase.Execute(context.Request.Context(), UpdatePostInput{
		ID:      entities.PostId(uriRequest.ID),
		Title:   bodyRequest.Title,
		Content: bodyRequest.Content,
	})
	if err != nil {
		context.Error(err)
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "Post updated successfully",
		"data":    post,
	})
}
