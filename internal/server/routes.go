package server

import (
	"net/http"

	createpost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/create_post"
	getpost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/get_post"
	listposts "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/list_posts"
	updatepost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/update_post"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine) {
	engine.GET("/healthy", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	engine.POST("/posts", createpost.Controller.Handle)
	engine.GET("/posts", listposts.Controller.Handle)
	engine.GET("/posts/:id", getpost.Controller.Handle)
	engine.PATCH("/posts/:id", updatepost.Controller.Handle)
}
