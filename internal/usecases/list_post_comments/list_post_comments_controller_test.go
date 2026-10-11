package listpostcomments_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/repositories"
	listpostcomments "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/list_post_comments"
	"github.com/gin-gonic/gin"
)

const validPostID = "6f1c2a3b-4d5e-4f60-8a7b-9c0d1e2f3a4b"

// listPostCommentsUseCaseMock implements listpostcomments.ListPostCommentsUseCase with a configurable Execute.
type listPostCommentsUseCaseMock struct {
	ExecuteFunc func(ctx context.Context, input listpostcomments.ListPostCommentsInput) (*repositories.PaginatedPostComments, error)
	calls       int
}

func (listPostCommentsUseCaseMock *listPostCommentsUseCaseMock) Execute(ctx context.Context, input listpostcomments.ListPostCommentsInput) (*repositories.PaginatedPostComments, error) {
	listPostCommentsUseCaseMock.calls++
	if listPostCommentsUseCaseMock.ExecuteFunc == nil {
		panic("listPostCommentsUseCaseMock.Execute called without ExecuteFunc")
	}
	return listPostCommentsUseCaseMock.ExecuteFunc(ctx, input)
}

// newListPostCommentsRequest builds a gin context for GET /posts/:id/comments with the given id param and raw query.
func newListPostCommentsRequest(test *testing.T, id string, query string) (*gin.Context, *httptest.ResponseRecorder) {
	test.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/posts/"+id+"/comments?"+query, nil)
	ginContext.Params = gin.Params{{Key: "id", Value: id}}
	return ginContext, recorder
}

func TestListPostCommentsController(test *testing.T) {

	test.Run("Handle", func(test *testing.T) {
		test.Run("Should call the use case with the ID from the path and no cursor nor limit when the query is empty", func(test *testing.T) {
			var received listpostcomments.ListPostCommentsInput
			listPostCommentsUseCase := &listPostCommentsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input listpostcomments.ListPostCommentsInput) (*repositories.PaginatedPostComments, error) {
					received = input
					return &repositories.PaginatedPostComments{}, nil
				},
			}
			listPostCommentsController := listpostcomments.NewListPostCommentsController(listpostcomments.NewListPostCommentsControllerInput{ListPostCommentsUseCase: listPostCommentsUseCase})
			ginContext, _ := newListPostCommentsRequest(test, validPostID, "")

			listPostCommentsController.Handle(ginContext)

			want := listpostcomments.ListPostCommentsInput{PostID: validPostID}
			if received != want {
				test.Errorf("Use case input = %+v, want %+v", received, want)
			}
		})

		test.Run("Should call the use case with the cursor and limit from the query", func(test *testing.T) {
			var received listpostcomments.ListPostCommentsInput
			listPostCommentsUseCase := &listPostCommentsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input listpostcomments.ListPostCommentsInput) (*repositories.PaginatedPostComments, error) {
					received = input
					return &repositories.PaginatedPostComments{}, nil
				},
			}
			listPostCommentsController := listpostcomments.NewListPostCommentsController(listpostcomments.NewListPostCommentsControllerInput{ListPostCommentsUseCase: listPostCommentsUseCase})
			ginContext, _ := newListPostCommentsRequest(test, validPostID, "cursor=abc&limit=10")

			listPostCommentsController.Handle(ginContext)

			if received.PostID != validPostID {
				test.Errorf("PostID = %q, want %q", received.PostID, validPostID)
			}
			if received.Cursor == nil || *received.Cursor != "abc" {
				test.Errorf("Cursor = %v, want %q", received.Cursor, "abc")
			}
			if received.Limit != 10 {
				test.Errorf("Limit = %d, want %d", received.Limit, 10)
			}
		})

		test.Run("Should pass the request context to the use case", func(test *testing.T) {
			type contextKey struct{}
			var receivedCtx context.Context
			listPostCommentsUseCase := &listPostCommentsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input listpostcomments.ListPostCommentsInput) (*repositories.PaginatedPostComments, error) {
					receivedCtx = ctx
					return &repositories.PaginatedPostComments{}, nil
				},
			}
			listPostCommentsController := listpostcomments.NewListPostCommentsController(listpostcomments.NewListPostCommentsControllerInput{ListPostCommentsUseCase: listPostCommentsUseCase})
			ginContext, _ := newListPostCommentsRequest(test, validPostID, "")
			ginContext.Request = ginContext.Request.WithContext(context.WithValue(ginContext.Request.Context(), contextKey{}, "request"))

			listPostCommentsController.Handle(ginContext)

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the use case to receive the request context, got %v", receivedCtx)
			}
		})

		test.Run("Should respond 200 with the page wrapped in the response envelope", func(test *testing.T) {
			createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
			nextCursor := "next-cursor"
			listPostCommentsUseCase := &listPostCommentsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input listpostcomments.ListPostCommentsInput) (*repositories.PaginatedPostComments, error) {
					return &repositories.PaginatedPostComments{
						Items:      []*entities.PostComment{{ID: "comment-id", PostID: "post-id", Comment: "comment", CreatedAt: createdAt, UpdatedAt: createdAt}},
						NextCursor: &nextCursor,
					}, nil
				},
			}
			listPostCommentsController := listpostcomments.NewListPostCommentsController(listpostcomments.NewListPostCommentsControllerInput{ListPostCommentsUseCase: listPostCommentsUseCase})
			ginContext, recorder := newListPostCommentsRequest(test, validPostID, "")

			listPostCommentsController.Handle(ginContext)

			if recorder.Code != http.StatusOK {
				test.Errorf("Status = %d, want %d", recorder.Code, http.StatusOK)
			}
			want := `{"data":{"items":[{"id":"comment-id","postId":"post-id","comment":"comment","createdAt":"2025-01-01T10:00:00Z","updatedAt":"2025-01-01T10:00:00Z"}],"nextCursor":"next-cursor"},"message":"Post comments listed successfully"}`
			if recorder.Body.String() != want {
				test.Errorf("Body = %s, want %s", recorder.Body.String(), want)
			}
			if len(ginContext.Errors) != 0 {
				test.Errorf("Expected no errors registered, got %v", ginContext.Errors)
			}
		})

		test.Run("Should register the use case error and not write a response", func(test *testing.T) {
			useCaseError := errors.New("use case failed")
			listPostCommentsUseCase := &listPostCommentsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input listpostcomments.ListPostCommentsInput) (*repositories.PaginatedPostComments, error) {
					return nil, useCaseError
				},
			}
			listPostCommentsController := listpostcomments.NewListPostCommentsController(listpostcomments.NewListPostCommentsControllerInput{ListPostCommentsUseCase: listPostCommentsUseCase})
			ginContext, _ := newListPostCommentsRequest(test, validPostID, "")

			listPostCommentsController.Handle(ginContext)

			lastError := ginContext.Errors.Last()
			if lastError == nil || !errors.Is(lastError.Err, useCaseError) {
				test.Errorf("Registered error = %v, want %v", lastError, useCaseError)
			}
			if ginContext.Writer.Written() {
				test.Error("Expected no response written, the error middleware is responsible for it")
			}
		})

		invalidRequests := []struct {
			name  string
			id    string
			query string
		}{
			{name: "id is empty", id: "", query: ""},
			{name: "id is not a UUID", id: "not-a-uuid", query: ""},
			{name: "limit is zero", id: validPostID, query: "limit=0"},
			{name: "limit is above the max", id: validPostID, query: "limit=101"},
			{name: "limit is not a number", id: validPostID, query: "limit=abc"},
		}
		for _, tc := range invalidRequests {
			test.Run("Should register an InvalidRequestParamsError and not call the use case when "+tc.name, func(test *testing.T) {
				listPostCommentsUseCase := &listPostCommentsUseCaseMock{}
				listPostCommentsController := listpostcomments.NewListPostCommentsController(listpostcomments.NewListPostCommentsControllerInput{ListPostCommentsUseCase: listPostCommentsUseCase})
				ginContext, _ := newListPostCommentsRequest(test, tc.id, tc.query)

				listPostCommentsController.Handle(ginContext)

				if listPostCommentsUseCase.calls != 0 {
					test.Errorf("Expected the use case not to be called, got %d calls", listPostCommentsUseCase.calls)
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
