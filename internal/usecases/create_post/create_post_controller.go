package createpost

import (
	"net/http"

	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/gin-gonic/gin"
)

type createPostRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type CreatePostController struct {
	createPostUseCase CreatePostUseCase
}

type NewCreatePostControllerInput struct {
	CreatePostUseCase CreatePostUseCase
}

func NewCreatePostController(input NewCreatePostControllerInput) *CreatePostController {
	return &CreatePostController{createPostUseCase: input.CreatePostUseCase}
}

// Handle creates a post.
//
//	@Summary	Create post
//	@Tags		posts
//	@Accept		json
//	@Produce	json
//	@Param		body	body		createPostRequest	true	"Post data"
//	@Success	201		{object}	entities.Post
//	@Failure	400		{object}	appErrors.AppError
//	@Failure	409		{object}	appErrors.AppError
//	@Failure	500		{object}	appErrors.AppError
//	@Router		/posts [post]
func (createPostController *CreatePostController) Handle(context *gin.Context) {
	const location = "CreatePostController.Handle"

	var request createPostRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.Error(appErrors.CreateInvalidRequestParamsError(appErrors.CreateInvalidRequestParamsErrorInput{
			Location:    location,
			OriginError: err,
		}))
		return
	}

	post, err := createPostController.createPostUseCase.Execute(context.Request.Context(), CreatePostInput{
		Title:   request.Title,
		Content: request.Content,
	})
	if err != nil {
		context.Error(err)
		return
	}

	context.JSON(http.StatusCreated, post)
}
