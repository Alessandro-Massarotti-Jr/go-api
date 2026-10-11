package server

import (
	"net/http"

	_ "github.com/Alessandro-Massarotti-Jr/go-api/docs"
	createpost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/create_post"
	createpostcomment "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/create_post_comment"
	deletepost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/delete_post"
	getpost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/get_post"
	listpostcomments "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/list_post_comments"
	listposts "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/list_posts"
	updatepost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/update_post"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(engine *gin.Engine) {
	engine.GET("/healthy", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"message": "ok", "data": nil})
	})

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	engine.POST("/posts", createpost.Controller.Handle)
	engine.GET("/posts", listposts.Controller.Handle)
	engine.GET("/posts/:id", getpost.Controller.Handle)
	engine.PATCH("/posts/:id", updatepost.Controller.Handle)
	engine.DELETE("/posts/:id", deletepost.Controller.Handle)

	engine.POST("/posts/:id/comments", createpostcomment.Controller.Handle)
	engine.GET("/posts/:id/comments", listpostcomments.Controller.Handle)
}
