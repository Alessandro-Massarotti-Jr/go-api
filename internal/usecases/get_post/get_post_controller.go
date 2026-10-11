package getpost

import (
	"net/http"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/gin-gonic/gin"
)

type getPostRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type GetPostController struct {
	getPostUseCase GetPostUseCase
}

type NewGetPostControllerInput struct {
	GetPostUseCase GetPostUseCase
}

func NewGetPostController(input NewGetPostControllerInput) *GetPostController {
	return &GetPostController{getPostUseCase: input.GetPostUseCase}
}

func (getPostController *GetPostController) Handle(context *gin.Context) {
	const location = "GetPostController.Handle"

	var request getPostRequest
	if err := context.ShouldBindUri(&request); err != nil {
		context.Error(appErrors.CreateInvalidRequestParamsError(appErrors.CreateInvalidRequestParamsErrorInput{
			Location:    location,
			OriginError: err,
		}))
		return
	}

	post, err := getPostController.getPostUseCase.Execute(context.Request.Context(), GetPostInput{
		ID: entities.PostId(request.ID),
	})
	if err != nil {
		context.Error(err)
		return
	}

	context.JSON(http.StatusOK, post)
}
