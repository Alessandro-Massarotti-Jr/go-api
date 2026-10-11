package updatepost_test

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
	updatepost "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/update_post"
	"github.com/gin-gonic/gin"
)

const validPostID = "6f1c2a3b-4d5e-4f60-8a7b-9c0d1e2f3a4b"

// updatePostUseCaseMock implements updatepost.UpdatePostUseCase with a configurable Execute.
type updatePostUseCaseMock struct {
	ExecuteFunc func(ctx context.Context, input updatepost.UpdatePostInput) (*entities.Post, error)
	calls       int
}

func (updatePostUseCaseMock *updatePostUseCaseMock) Execute(ctx context.Context, input updatepost.UpdatePostInput) (*entities.Post, error) {
	updatePostUseCaseMock.calls++
	if updatePostUseCaseMock.ExecuteFunc == nil {
		panic("updatePostUseCaseMock.Execute called without ExecuteFunc")
	}
	return updatePostUseCaseMock.ExecuteFunc(ctx, input)
}

// newUpdatePostRequest builds a gin context for PATCH /posts/:id with the given id param and JSON body.
func newUpdatePostRequest(test *testing.T, id string, body string) (*gin.Context, *httptest.ResponseRecorder) {
	test.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodPatch, "/posts/"+id, strings.NewReader(body))
	ginContext.Request.Header.Set("Content-Type", "application/json")
	ginContext.Params = gin.Params{{Key: "id", Value: id}}
	return ginContext, recorder
}

func TestUpdatePostController(test *testing.T) {

	test.Run("Handle", func(test *testing.T) {
		test.Run("Should call the use case with the ID from the path and the title and content from the body", func(test *testing.T) {
			var received updatepost.UpdatePostInput
			updatePostUseCase := &updatePostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input updatepost.UpdatePostInput) (*entities.Post, error) {
					received = input
					return &entities.Post{}, nil
				},
			}
			updatePostController := updatepost.NewUpdatePostController(updatepost.NewUpdatePostControllerInput{UpdatePostUseCase: updatePostUseCase})
			ginContext, _ := newUpdatePostRequest(test, validPostID, `{"title":"New title","content":"New content"}`)

			updatePostController.Handle(ginContext)

			if received.ID != validPostID {
				test.Errorf("ID = %q, want %q", received.ID, validPostID)
			}
			if received.Title == nil || *received.Title != "New title" {
				test.Errorf("Title = %v, want %q", received.Title, "New title")
			}
			if received.Content == nil || *received.Content != "New content" {
				test.Errorf("Content = %v, want %q", received.Content, "New content")
			}
		})

		test.Run("Should call the use case with nil for the fields missing in the body", func(test *testing.T) {
			var received updatepost.UpdatePostInput
			updatePostUseCase := &updatePostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input updatepost.UpdatePostInput) (*entities.Post, error) {
					received = input
					return &entities.Post{}, nil
				},
			}
			updatePostController := updatepost.NewUpdatePostController(updatepost.NewUpdatePostControllerInput{UpdatePostUseCase: updatePostUseCase})
			ginContext, _ := newUpdatePostRequest(test, validPostID, `{"title":"New title"}`)

			updatePostController.Handle(ginContext)

			if received.Title == nil || *received.Title != "New title" {
				test.Errorf("Title = %v, want %q", received.Title, "New title")
			}
			if received.Content != nil {
				test.Errorf("Content = %v, want nil", received.Content)
			}
		})

		test.Run("Should pass the request context to the use case", func(test *testing.T) {
			type contextKey struct{}
			var receivedCtx context.Context
			updatePostUseCase := &updatePostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input updatepost.UpdatePostInput) (*entities.Post, error) {
					receivedCtx = ctx
					return &entities.Post{}, nil
				},
			}
			updatePostController := updatepost.NewUpdatePostController(updatepost.NewUpdatePostControllerInput{UpdatePostUseCase: updatePostUseCase})
			ginContext, _ := newUpdatePostRequest(test, validPostID, `{"title":"title"}`)
			ginContext.Request = ginContext.Request.WithContext(context.WithValue(ginContext.Request.Context(), contextKey{}, "request"))

			updatePostController.Handle(ginContext)

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the use case to receive the request context, got %v", receivedCtx)
			}
		})

		test.Run("Should respond 200 with the updated post wrapped in the response envelope", func(test *testing.T) {
			createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
			updatedAt := time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)
			updatePostUseCase := &updatePostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input updatepost.UpdatePostInput) (*entities.Post, error) {
					return &entities.Post{ID: "post-id", Title: "title", Content: "content", CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
				},
			}
			updatePostController := updatepost.NewUpdatePostController(updatepost.NewUpdatePostControllerInput{UpdatePostUseCase: updatePostUseCase})
			ginContext, recorder := newUpdatePostRequest(test, validPostID, `{"title":"title"}`)

			updatePostController.Handle(ginContext)

			if recorder.Code != http.StatusOK {
				test.Errorf("Status = %d, want %d", recorder.Code, http.StatusOK)
			}
			want := `{"data":{"id":"post-id","title":"title","content":"content","createdAt":"2025-01-01T10:00:00Z","updatedAt":"2025-01-02T10:00:00Z"},"message":"Post updated successfully"}`
			if recorder.Body.String() != want {
				test.Errorf("Body = %s, want %s", recorder.Body.String(), want)
			}
			if len(ginContext.Errors) != 0 {
				test.Errorf("Expected no errors registered, got %v", ginContext.Errors)
			}
		})

		test.Run("Should register the use case error and not write a response", func(test *testing.T) {
			useCaseError := errors.New("use case failed")
			updatePostUseCase := &updatePostUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input updatepost.UpdatePostInput) (*entities.Post, error) {
					return nil, useCaseError
				},
			}
			updatePostController := updatepost.NewUpdatePostController(updatepost.NewUpdatePostControllerInput{UpdatePostUseCase: updatePostUseCase})
			ginContext, _ := newUpdatePostRequest(test, validPostID, `{"title":"title"}`)

			updatePostController.Handle(ginContext)

			lastError := ginContext.Errors.Last()
			if lastError == nil || !errors.Is(lastError.Err, useCaseError) {
				test.Errorf("Registered error = %v, want %v", lastError, useCaseError)
			}
			if ginContext.Writer.Written() {
				test.Error("Expected no response written, the error middleware is responsible for it")
			}
		})

		invalidRequests := []struct {
			name string
			id   string
			body string
		}{
			{name: "id is empty", id: "", body: `{"title":"title"}`},
			{name: "id is not a UUID", id: "not-a-uuid", body: `{"title":"title"}`},
			{name: "malformed JSON", id: validPostID, body: `{"title":`},
			{name: "empty body", id: validPostID, body: ``},
			{name: "empty object", id: validPostID, body: `{}`},
			{name: "empty title", id: validPostID, body: `{"title":""}`},
			{name: "empty content", id: validPostID, body: `{"content":""}`},
			{name: "title with wrong type", id: validPostID, body: `{"title":1}`},
		}
		for _, tc := range invalidRequests {
			test.Run("Should register an InvalidRequestParamsError and not call the use case when "+tc.name, func(test *testing.T) {
				updatePostUseCase := &updatePostUseCaseMock{}
				updatePostController := updatepost.NewUpdatePostController(updatepost.NewUpdatePostControllerInput{UpdatePostUseCase: updatePostUseCase})
				ginContext, _ := newUpdatePostRequest(test, tc.id, tc.body)

				updatePostController.Handle(ginContext)

				if updatePostUseCase.calls != 0 {
					test.Errorf("Expected the use case not to be called, got %d calls", updatePostUseCase.calls)
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
