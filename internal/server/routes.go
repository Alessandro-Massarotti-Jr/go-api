package server

import (
	"net/http"

	createpost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/create_post"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine) {
	engine.GET("/healthy", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	engine.POST("/posts", createpost.Controller.Handle)
}
