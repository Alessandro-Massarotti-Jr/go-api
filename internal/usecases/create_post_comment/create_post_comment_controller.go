package createpostcomment

import (
	"net/http"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/gin-gonic/gin"
)

type createPostCommentUriRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type createPostCommentBodyRequest struct {
	Comment string `json:"comment" binding:"required"`
}

type CreatePostCommentController struct {
	createPostCommentUseCase CreatePostCommentUseCase
}

type NewCreatePostCommentControllerInput struct {
	CreatePostCommentUseCase CreatePostCommentUseCase
}

func NewCreatePostCommentController(input NewCreatePostCommentControllerInput) *CreatePostCommentController {
	return &CreatePostCommentController{createPostCommentUseCase: input.CreatePostCommentUseCase}
}

// Handle adds a comment to a post.
//
//	@Summary	Create post comment
//	@Tags		post comments
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string							true	"Post id (UUID)"	format(uuid)
//	@Param		body	body		createPostCommentBodyRequest	true	"Comment data"
//	@Success	201		{object}	object{message=string,data=entities.PostComment}
//	@Failure	400		{object}	object{error=appErrors.AppError}
//	@Failure	404		{object}	object{error=appErrors.AppError}
//	@Failure	500		{object}	object{error=appErrors.AppError}
//	@Router		/posts/{id}/comments [post]
func (createPostCommentController *CreatePostCommentController) Handle(context *gin.Context) {
	const location = "CreatePostCommentController.Handle"

	var uriRequest createPostCommentUriRequest
	if err := context.ShouldBindUri(&uriRequest); err != nil {
		context.Error(appErrors.CreateInvalidRequestParamsError(appErrors.CreateInvalidRequestParamsErrorInput{
			Location:    location,
			OriginError: err,
		}))
		return
	}

	var bodyRequest createPostCommentBodyRequest
	if err := context.ShouldBindJSON(&bodyRequest); err != nil {
		context.Error(appErrors.CreateInvalidRequestParamsError(appErrors.CreateInvalidRequestParamsErrorInput{
			Location:    location,
			OriginError: err,
		}))
		return
	}

	postComment, err := createPostCommentController.createPostCommentUseCase.Execute(context.Request.Context(), CreatePostCommentInput{
		PostID:  entities.PostId(uriRequest.ID),
		Comment: bodyRequest.Comment,
	})
	if err != nil {
		context.Error(err)
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"message": "Post comment created successfully",
		"data":    postComment,
	})
}
