package createpost_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	createpost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/create_post"
	"github.com/gin-gonic/gin"
)

// createPostUseCaseMock implements createpost.CreatePostUseCase with a configurable Execute.
type createPostUseCaseMock struct {
	ExecuteFunc func(ctx context.Context, input createpost.CreatePostInput) (*entities.Post, error)
	calls       int
}

func (createPostUseCaseMock *createPostUseCaseMock) Execute(ctx context.Context, input createpost.CreatePostInput) (*entities.Post, error) {
	createPostUseCaseMock.calls++
	if createPostUseCaseMock.ExecuteFunc == nil {
		panic("createPostUseCaseMock.Execute called without ExecuteFunc")
	}
	return createPostUseCaseMock.ExecuteFunc(ctx, input)
}

// newPostRequest builds a gin context for POST /posts with the given JSON body.
func newPostRequest(test *testing.T, body string) (*gin.Context, *httptest.ResponseRecorder) {
	test.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(body))
	ginContext.Request.Header.Set("Content-Type", "application/json")
	return ginContext, recorder
}

func TestCreatePostController(test *testing.T) {

	test.Run("Handle", func(test *testing.T) {
		test.Run("Should call the use case with the title and content from the body", func(test *testing.T) {
			var received createpost.CreatePostInput
			createPostUseCase := &createPostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input createpost.CreatePostInput) (*entities.Post, error) {
					received = input
					return &entities.Post{}, nil
				},
			}
			createPostController := createpost.NewCreatePostController(createpost.NewCreatePostControllerInput{CreatePostUseCase: createPostUseCase})
			ginContext, _ := newPostRequest(test, `{"title":"My title","content":"My content"}`)

			createPostController.Handle(ginContext)

			want := createpost.CreatePostInput{Title: "My title", Content: "My content"}
			if received != want {
				test.Errorf("Use case input = %+v, want %+v", received, want)
			}
		})

		test.Run("Should pass the request context to the use case", func(test *testing.T) {
			type contextKey struct{}
			var receivedCtx context.Context
			createPostUseCase := &createPostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input createpost.CreatePostInput) (*entities.Post, error) {
					receivedCtx = ctx
					return &entities.Post{}, nil
				},
			}
			createPostController := createpost.NewCreatePostController(createpost.NewCreatePostControllerInput{CreatePostUseCase: createPostUseCase})
			ginContext, _ := newPostRequest(test, `{"title":"title","content":"content"}`)
			ginContext.Request = ginContext.Request.WithContext(context.WithValue(ginContext.Request.Context(), contextKey{}, "request"))

			createPostController.Handle(ginContext)

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the use case to receive the request context, got %v", receivedCtx)
			}
		})

		test.Run("Should respond 201 with the created post as JSON", func(test *testing.T) {
			createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
			createPostUseCase := &createPostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input createpost.CreatePostInput) (*entities.Post, error) {
					return &entities.Post{ID: "post-id", Title: "title", Content: "content", CreatedAt: createdAt, UpdatedAt: createdAt}, nil
				},
			}
			createPostController := createpost.NewCreatePostController(createpost.NewCreatePostControllerInput{CreatePostUseCase: createPostUseCase})
			ginContext, recorder := newPostRequest(test, `{"title":"title","content":"content"}`)

			createPostController.Handle(ginContext)

			if recorder.Code != http.StatusCreated {
				test.Errorf("Status = %d, want %d", recorder.Code, http.StatusCreated)
			}
			want := `{"id":"post-id","title":"title","content":"content","createdAt":"2025-01-01T10:00:00Z","updatedAt":"2025-01-01T10:00:00Z"}`
			if recorder.Body.String() != want {
				test.Errorf("Body = %s, want %s", recorder.Body.String(), want)
			}
			if len(ginContext.Errors) != 0 {
				test.Errorf("Expected no errors registered, got %v", ginContext.Errors)
			}
		})

		test.Run("Should register the use case error and not write a response", func(test *testing.T) {
			useCaseError := errors.New("use case failed")
			createPostUseCase := &createPostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input createpost.CreatePostInput) (*entities.Post, error) {
					return nil, useCaseError
				},
			}
			createPostController := createpost.NewCreatePostController(createpost.NewCreatePostControllerInput{CreatePostUseCase: createPostUseCase})
			ginContext, _ := newPostRequest(test, `{"title":"title","content":"content"}`)

			createPostController.Handle(ginContext)

			lastError := ginContext.Errors.Last()
			if lastError == nil || !errors.Is(lastError.Err, useCaseError) {
				test.Errorf("Registered error = %v, want %v", lastError, useCaseError)
			}
			if ginContext.Writer.Written() {
				test.Error("Expected no response written, the error middleware is responsible for it")
			}
		})

		invalidBodies := []struct {
			name string
			body string
		}{
			{name: "malformed JSON", body: `{"title":`},
			{name: "empty body", body: ``},
			{name: "empty object", body: `{}`},
			{name: "missing title", body: `{"content":"content"}`},
			{name: "missing content", body: `{"title":"title"}`},
			{name: "empty title", body: `{"title":"","content":"content"}`},
			{name: "empty content", body: `{"title":"title","content":""}`},
			{name: "title with wrong type", body: `{"title":1,"content":"content"}`},
		}
		for _, tc := range invalidBodies {
			test.Run("Should register an InvalidRequestParamsError and not call the use case when "+tc.name, func(test *testing.T) {
				createPostUseCase := &createPostUseCaseMock{}
				createPostController := createpost.NewCreatePostController(createpost.NewCreatePostControllerInput{CreatePostUseCase: createPostUseCase})
				ginContext, _ := newPostRequest(test, tc.body)

				createPostController.Handle(ginContext)

				if createPostUseCase.calls != 0 {
					test.Errorf("Expected the use case not to be called, got %d calls", createPostUseCase.calls)
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
