package createpostcomment_test

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
	createpostcomment "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/create_post_comment"
	"github.com/gin-gonic/gin"
)

const validPostID = "6f1c2a3b-4d5e-4f60-8a7b-9c0d1e2f3a4b"

// createPostCommentUseCaseMock implements createpostcomment.CreatePostCommentUseCase with a configurable Execute.
type createPostCommentUseCaseMock struct {
	ExecuteFunc func(ctx context.Context, input createpostcomment.CreatePostCommentInput) (*entities.PostComment, error)
	calls       int
}

func (createPostCommentUseCaseMock *createPostCommentUseCaseMock) Execute(ctx context.Context, input createpostcomment.CreatePostCommentInput) (*entities.PostComment, error) {
	createPostCommentUseCaseMock.calls++
	if createPostCommentUseCaseMock.ExecuteFunc == nil {
		panic("createPostCommentUseCaseMock.Execute called without ExecuteFunc")
	}
	return createPostCommentUseCaseMock.ExecuteFunc(ctx, input)
}

// newCreatePostCommentRequest builds a gin context for POST /posts/:id/comments with the given id param and JSON body.
func newCreatePostCommentRequest(test *testing.T, id string, body string) (*gin.Context, *httptest.ResponseRecorder) {
	test.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodPost, "/posts/"+id+"/comments", strings.NewReader(body))
	ginContext.Request.Header.Set("Content-Type", "application/json")
	ginContext.Params = gin.Params{{Key: "id", Value: id}}
	return ginContext, recorder
}

func TestCreatePostCommentController(test *testing.T) {

	test.Run("Handle", func(test *testing.T) {
		test.Run("Should call the use case with the ID from the path and the comment from the body", func(test *testing.T) {
			var received createpostcomment.CreatePostCommentInput
			createPostCommentUseCase := &createPostCommentUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input createpostcomment.CreatePostCommentInput) (*entities.PostComment, error) {
					received = input
					return &entities.PostComment{}, nil
				},
			}
			createPostCommentController := createpostcomment.NewCreatePostCommentController(createpostcomment.NewCreatePostCommentControllerInput{CreatePostCommentUseCase: createPostCommentUseCase})
			ginContext, _ := newCreatePostCommentRequest(test, validPostID, `{"comment":"My comment"}`)

			createPostCommentController.Handle(ginContext)

			want := createpostcomment.CreatePostCommentInput{PostID: validPostID, Comment: "My comment"}
			if received != want {
				test.Errorf("Use case input = %+v, want %+v", received, want)
			}
		})

		test.Run("Should pass the request context to the use case", func(test *testing.T) {
			type contextKey struct{}
			var receivedCtx context.Context
			createPostCommentUseCase := &createPostCommentUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input createpostcomment.CreatePostCommentInput) (*entities.PostComment, error) {
					receivedCtx = ctx
					return &entities.PostComment{}, nil
				},
			}
			createPostCommentController := createpostcomment.NewCreatePostCommentController(createpostcomment.NewCreatePostCommentControllerInput{CreatePostCommentUseCase: createPostCommentUseCase})
			ginContext, _ := newCreatePostCommentRequest(test, validPostID, `{"comment":"comment"}`)
			ginContext.Request = ginContext.Request.WithContext(context.WithValue(ginContext.Request.Context(), contextKey{}, "request"))

			createPostCommentController.Handle(ginContext)

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the use case to receive the request context, got %v", receivedCtx)
			}
		})

		test.Run("Should respond 201 with the created comment wrapped in the response envelope", func(test *testing.T) {
			createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
			createPostCommentUseCase := &createPostCommentUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input createpostcomment.CreatePostCommentInput) (*entities.PostComment, error) {
					return &entities.PostComment{ID: "comment-id", PostID: "post-id", Comment: "comment", CreatedAt: createdAt, UpdatedAt: createdAt}, nil
				},
			}
			createPostCommentController := createpostcomment.NewCreatePostCommentController(createpostcomment.NewCreatePostCommentControllerInput{CreatePostCommentUseCase: createPostCommentUseCase})
			ginContext, recorder := newCreatePostCommentRequest(test, validPostID, `{"comment":"comment"}`)

			createPostCommentController.Handle(ginContext)

			if recorder.Code != http.StatusCreated {
				test.Errorf("Status = %d, want %d", recorder.Code, http.StatusCreated)
			}
			want := `{"data":{"id":"comment-id","postId":"post-id","comment":"comment","createdAt":"2025-01-01T10:00:00Z","updatedAt":"2025-01-01T10:00:00Z"},"message":"Post comment created successfully"}`
			if recorder.Body.String() != want {
				test.Errorf("Body = %s, want %s", recorder.Body.String(), want)
			}
			if len(ginContext.Errors) != 0 {
				test.Errorf("Expected no errors registered, got %v", ginContext.Errors)
			}
		})

		test.Run("Should register the use case error and not write a response", func(test *testing.T) {
			useCaseError := errors.New("use case failed")
			createPostCommentUseCase := &createPostCommentUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input createpostcomment.CreatePostCommentInput) (*entities.PostComment, error) {
					return nil, useCaseError
				},
			}
			createPostCommentController := createpostcomment.NewCreatePostCommentController(createpostcomment.NewCreatePostCommentControllerInput{CreatePostCommentUseCase: createPostCommentUseCase})
			ginContext, _ := newCreatePostCommentRequest(test, validPostID, `{"comment":"comment"}`)

			createPostCommentController.Handle(ginContext)

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
			{name: "id is empty", id: "", body: `{"comment":"comment"}`},
			{name: "id is not a UUID", id: "not-a-uuid", body: `{"comment":"comment"}`},
			{name: "malformed JSON", id: validPostID, body: `{"comment":`},
			{name: "empty body", id: validPostID, body: ``},
			{name: "missing comment", id: validPostID, body: `{}`},
			{name: "empty comment", id: validPostID, body: `{"comment":""}`},
			{name: "comment with wrong type", id: validPostID, body: `{"comment":1}`},
		}
		for _, tc := range invalidRequests {
			test.Run("Should register an InvalidRequestParamsError and not call the use case when "+tc.name, func(test *testing.T) {
				createPostCommentUseCase := &createPostCommentUseCaseMock{}
				createPostCommentController := createpostcomment.NewCreatePostCommentController(createpostcomment.NewCreatePostCommentControllerInput{CreatePostCommentUseCase: createPostCommentUseCase})
				ginContext, _ := newCreatePostCommentRequest(test, tc.id, tc.body)

				createPostCommentController.Handle(ginContext)

				if createPostCommentUseCase.calls != 0 {
					test.Errorf("Expected the use case not to be called, got %d calls", createPostCommentUseCase.calls)
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
