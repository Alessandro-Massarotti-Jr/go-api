package getpost_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	getpost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/get_post"
	"github.com/gin-gonic/gin"
)

const validPostID = "6f1c2a3b-4d5e-4f60-8a7b-9c0d1e2f3a4b"

// getPostUseCaseMock implements getpost.GetPostUseCase with a configurable Execute.
type getPostUseCaseMock struct {
	ExecuteFunc func(ctx context.Context, input getpost.GetPostInput) (*entities.Post, error)
	calls       int
}

func (getPostUseCaseMock *getPostUseCaseMock) Execute(ctx context.Context, input getpost.GetPostInput) (*entities.Post, error) {
	getPostUseCaseMock.calls++
	if getPostUseCaseMock.ExecuteFunc == nil {
		panic("getPostUseCaseMock.Execute called without ExecuteFunc")
	}
	return getPostUseCaseMock.ExecuteFunc(ctx, input)
}

// newGetPostRequest builds a gin context for GET /posts/:id with the given id param.
func newGetPostRequest(test *testing.T, id string) (*gin.Context, *httptest.ResponseRecorder) {
	test.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/posts/"+id, nil)
	ginContext.Params = gin.Params{{Key: "id", Value: id}}
	return ginContext, recorder
}

func TestGetPostController(test *testing.T) {

	test.Run("Handle", func(test *testing.T) {
		test.Run("Should call the use case with the ID from the path", func(test *testing.T) {
			var received getpost.GetPostInput
			getPostUseCase := &getPostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input getpost.GetPostInput) (*entities.Post, error) {
					received = input
					return &entities.Post{}, nil
				},
			}
			getPostController := getpost.NewGetPostController(getpost.NewGetPostControllerInput{GetPostUseCase: getPostUseCase})
			ginContext, _ := newGetPostRequest(test, validPostID)

			getPostController.Handle(ginContext)

			want := getpost.GetPostInput{ID: validPostID}
			if received != want {
				test.Errorf("Use case input = %+v, want %+v", received, want)
			}
		})

		test.Run("Should pass the request context to the use case", func(test *testing.T) {
			type contextKey struct{}
			var receivedCtx context.Context
			getPostUseCase := &getPostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input getpost.GetPostInput) (*entities.Post, error) {
					receivedCtx = ctx
					return &entities.Post{}, nil
				},
			}
			getPostController := getpost.NewGetPostController(getpost.NewGetPostControllerInput{GetPostUseCase: getPostUseCase})
			ginContext, _ := newGetPostRequest(test, validPostID)
			ginContext.Request = ginContext.Request.WithContext(context.WithValue(ginContext.Request.Context(), contextKey{}, "request"))

			getPostController.Handle(ginContext)

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the use case to receive the request context, got %v", receivedCtx)
			}
		})

		test.Run("Should respond 200 with the post wrapped in the response envelope", func(test *testing.T) {
			createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
			getPostUseCase := &getPostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input getpost.GetPostInput) (*entities.Post, error) {
					return &entities.Post{ID: "post-id", Title: "title", Content: "content", CreatedAt: createdAt, UpdatedAt: createdAt}, nil
				},
			}
			getPostController := getpost.NewGetPostController(getpost.NewGetPostControllerInput{GetPostUseCase: getPostUseCase})
			ginContext, recorder := newGetPostRequest(test, validPostID)

			getPostController.Handle(ginContext)

			if recorder.Code != http.StatusOK {
				test.Errorf("Status = %d, want %d", recorder.Code, http.StatusOK)
			}
			want := `{"data":{"id":"post-id","title":"title","content":"content","totalComments":0,"createdAt":"2025-01-01T10:00:00Z","updatedAt":"2025-01-01T10:00:00Z"},"message":"Post retrieved successfully"}`
			if recorder.Body.String() != want {
				test.Errorf("Body = %s, want %s", recorder.Body.String(), want)
			}
			if len(ginContext.Errors) != 0 {
				test.Errorf("Expected no errors registered, got %v", ginContext.Errors)
			}
		})

		test.Run("Should register the use case error and not write a response", func(test *testing.T) {
			useCaseError := errors.New("use case failed")
			getPostUseCase := &getPostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input getpost.GetPostInput) (*entities.Post, error) {
					return nil, useCaseError
				},
			}
			getPostController := getpost.NewGetPostController(getpost.NewGetPostControllerInput{GetPostUseCase: getPostUseCase})
			ginContext, _ := newGetPostRequest(test, validPostID)

			getPostController.Handle(ginContext)

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
				getPostUseCase := &getPostUseCaseMock{}
				getPostController := getpost.NewGetPostController(getpost.NewGetPostControllerInput{GetPostUseCase: getPostUseCase})
				ginContext, _ := newGetPostRequest(test, tc.id)

				getPostController.Handle(ginContext)

				if getPostUseCase.calls != 0 {
					test.Errorf("Expected the use case not to be called, got %d calls", getPostUseCase.calls)
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
