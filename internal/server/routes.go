package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine) {
	engine.GET("/healthy", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
}
