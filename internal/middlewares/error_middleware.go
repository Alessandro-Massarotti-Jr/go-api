package middlewares

import (
	"errors"

	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/gin-gonic/gin"
)

// ErrorHandler turns the last error registered by a handler (via context.Error)
// into the JSON response. AppErrors keep their HTTP status; any other error becomes
// a generic 500 so internal details are not exposed.
func ErrorHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Next()
		lastError := context.Errors.Last()
		if lastError == nil {
			return
		}
		if appError, isAppError := errors.AsType[*appErrors.AppError](lastError.Err); isAppError {
			context.JSON(appError.HTTPStatus, appError)
			return
		}

		genericError := appErrors.Create(appErrors.CreateAppErrorInput{
			Location:    context.Request.Method + " " + context.FullPath(),
			OriginError: lastError.Err,
		})

		context.JSON(genericError.HTTPStatus, genericError)
	}
}
