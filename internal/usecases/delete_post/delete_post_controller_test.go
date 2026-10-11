package deletepost_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	deletepost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/delete_post"
	"github.com/gin-gonic/gin"
)

const validPostID = "6f1c2a3b-4d5e-4f60-8a7b-9c0d1e2f3a4b"

// deletePostUseCaseMock implements deletepost.DeletePostUseCase with a configurable Execute.
type deletePostUseCaseMock struct {
	ExecuteFunc func(ctx context.Context, input deletepost.DeletePostInput) error
	calls       int
}

func (deletePostUseCaseMock *deletePostUseCaseMock) Execute(ctx context.Context, input deletepost.DeletePostInput) error {
	deletePostUseCaseMock.calls++
	if deletePostUseCaseMock.ExecuteFunc == nil {
		panic("deletePostUseCaseMock.Execute called without ExecuteFunc")
	}
	return deletePostUseCaseMock.ExecuteFunc(ctx, input)
}

// newDeletePostRequest builds a gin context for DELETE /posts/:id with the given id param.
func newDeletePostRequest(test *testing.T, id string) (*gin.Context, *httptest.ResponseRecorder) {
	test.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodDelete, "/posts/"+id, nil)
	ginContext.Params = gin.Params{{Key: "id", Value: id}}
	return ginContext, recorder
}

func TestDeletePostController(test *testing.T) {

	test.Run("Handle", func(test *testing.T) {
		test.Run("Should call the use case with the ID from the path", func(test *testing.T) {
			var received deletepost.DeletePostInput
			deletePostUseCase := &deletePostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input deletepost.DeletePostInput) error {
					received = input
					return nil
				},
			}
			deletePostController := deletepost.NewDeletePostController(deletepost.NewDeletePostControllerInput{DeletePostUseCase: deletePostUseCase})
			ginContext, _ := newDeletePostRequest(test, validPostID)

			deletePostController.Handle(ginContext)

			want := deletepost.DeletePostInput{ID: validPostID}
			if received != want {
				test.Errorf("Use case input = %+v, want %+v", received, want)
			}
		})

		test.Run("Should pass the request context to the use case", func(test *testing.T) {
			type contextKey struct{}
			var receivedCtx context.Context
			deletePostUseCase := &deletePostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input deletepost.DeletePostInput) error {
					receivedCtx = ctx
					return nil
				},
			}
			deletePostController := deletepost.NewDeletePostController(deletepost.NewDeletePostControllerInput{DeletePostUseCase: deletePostUseCase})
			ginContext, _ := newDeletePostRequest(test, validPostID)
			ginContext.Request = ginContext.Request.WithContext(context.WithValue(ginContext.Request.Context(), contextKey{}, "request"))

			deletePostController.Handle(ginContext)

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the use case to receive the request context, got %v", receivedCtx)
			}
		})

		test.Run("Should respond 200 with the deletion message", func(test *testing.T) {
			deletePostUseCase := &deletePostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input deletepost.DeletePostInput) error {
					return nil
				},
			}
			deletePostController := deletepost.NewDeletePostController(deletepost.NewDeletePostControllerInput{DeletePostUseCase: deletePostUseCase})
			ginContext, recorder := newDeletePostRequest(test, validPostID)

			deletePostController.Handle(ginContext)

			if recorder.Code != http.StatusOK {
				test.Errorf("Status = %d, want %d", recorder.Code, http.StatusOK)
			}
			want := `{"data":null,"message":"Post deleted successfully"}`
			if recorder.Body.String() != want {
				test.Errorf("Body = %s, want %s", recorder.Body.String(), want)
			}
			if len(ginContext.Errors) != 0 {
				test.Errorf("Expected no errors registered, got %v", ginContext.Errors)
			}
		})

		test.Run("Should register the use case error and not write a response", func(test *testing.T) {
			useCaseError := errors.New("use case failed")
			deletePostUseCase := &deletePostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input deletepost.DeletePostInput) error {
					return useCaseError
				},
			}
			deletePostController := deletepost.NewDeletePostController(deletepost.NewDeletePostControllerInput{DeletePostUseCase: deletePostUseCase})
			ginContext, _ := newDeletePostRequest(test, validPostID)

			deletePostController.Handle(ginContext)

			lastError := ginContext.Errors.Last()
			if lastError == nil || !errors.Is(lastError.Err, useCaseError) {
				test.Errorf("Registered error = %v, want %v", lastError, useCaseError)
			}
			if ginContext.Writer.Written() {
				test.Error("Expected no response written, the error middleware is responsible for it")
			}
		})

		invalidIDs := []struct {
			name string
			id   string
		}{
			{name: "id is empty", id: ""},
			{name: "id is not a UUID", id: "not-a-uuid"},
		}
		for _, tc := range invalidIDs {
			test.Run("Should register an InvalidRequestParamsError and not call the use case when "+tc.name, func(test *testing.T) {
				deletePostUseCase := &deletePostUseCaseMock{}
				deletePostController := deletepost.NewDeletePostController(deletepost.NewDeletePostControllerInput{DeletePostUseCase: deletePostUseCase})
				ginContext, _ := newDeletePostRequest(test, tc.id)

				deletePostController.Handle(ginContext)

				if deletePostUseCase.calls != 0 {
					test.Errorf("Expected the use case not to be called, got %d calls", deletePostUseCase.calls)
				}

				lastError := ginContext.Errors.Last()
				if lastError == nil {
					test.Fatal("Expected an error registered, got none")
				}
				appError, isAppError := errors.AsType[*appErrors.AppError](lastError.Err)
				if !isAppError {
					test.Fatalf("Registered error = %v, want an AppError", lastError.Err)
				}
				if appError.Name != "InvalidRequestParamsError" {
					test.Errorf("Name = %q, want %q", appError.Name, "InvalidRequestParamsError")
				}
				if appError.HTTPStatus != http.StatusBadRequest {
					test.Errorf("HTTPStatus = %d, want %d", appError.HTTPStatus, http.StatusBadRequest)
				}
				if ginContext.Writer.Written() {
					test.Error("Expected no response written, the error middleware is responsible for it")
				}
			})
		}
	})

}
