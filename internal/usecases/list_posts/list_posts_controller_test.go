package listposts_test

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
	listposts "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/list_posts"
	"github.com/gin-gonic/gin"
)

// listPostsUseCaseMock implements listposts.ListPostsUseCase with a configurable Execute.
type listPostsUseCaseMock struct {
	ExecuteFunc func(ctx context.Context, input listposts.ListPostsInput) (*repositories.PaginatedPosts, error)
	calls       int
}

func (listPostsUseCaseMock *listPostsUseCaseMock) Execute(ctx context.Context, input listposts.ListPostsInput) (*repositories.PaginatedPosts, error) {
	listPostsUseCaseMock.calls++
	if listPostsUseCaseMock.ExecuteFunc == nil {
		panic("listPostsUseCaseMock.Execute called without ExecuteFunc")
	}
	return listPostsUseCaseMock.ExecuteFunc(ctx, input)
}

// newListPostsRequest builds a gin context for GET /posts with the given query string.
func newListPostsRequest(test *testing.T, query string) (*gin.Context, *httptest.ResponseRecorder) {
	test.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/posts"+query, nil)
	return ginContext, recorder
}

func TestListPostsController(test *testing.T) {

	test.Run("Handle", func(test *testing.T) {
		test.Run("Should call the use case with the cursor and limit from the query", func(test *testing.T) {
			var received listposts.ListPostsInput
			listPostsUseCase := &listPostsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input listposts.ListPostsInput) (*repositories.PaginatedPosts, error) {
					received = input
					return &repositories.PaginatedPosts{}, nil
				},
			}
			listPostsController := listposts.NewListPostsController(listposts.NewListPostsControllerInput{ListPostsUseCase: listPostsUseCase})
			ginContext, _ := newListPostsRequest(test, "?cursor=next-cursor&limit=10")

			listPostsController.Handle(ginContext)

			if received.Cursor == nil || *received.Cursor != "next-cursor" {
				test.Errorf("Cursor = %v, want %q", received.Cursor, "next-cursor")
			}
			if received.Limit != 10 {
				test.Errorf("Limit = %d, want %d", received.Limit, 10)
			}
		})

		test.Run("Should call the use case with a nil cursor and zero limit when the query is empty", func(test *testing.T) {
			var received listposts.ListPostsInput
			listPostsUseCase := &listPostsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input listposts.ListPostsInput) (*repositories.PaginatedPosts, error) {
					received = input
					return &repositories.PaginatedPosts{}, nil
				},
			}
			listPostsController := listposts.NewListPostsController(listposts.NewListPostsControllerInput{ListPostsUseCase: listPostsUseCase})
			ginContext, _ := newListPostsRequest(test, "")

			listPostsController.Handle(ginContext)

			if received.Cursor != nil {
				test.Errorf("Cursor = %v, want nil", received.Cursor)
			}
			if received.Limit != 0 {
				test.Errorf("Limit = %d, want %d", received.Limit, 0)
			}
		})

		test.Run("Should pass the request context to the use case", func(test *testing.T) {
			type contextKey struct{}
			var receivedCtx context.Context
			listPostsUseCase := &listPostsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input listposts.ListPostsInput) (*repositories.PaginatedPosts, error) {
					receivedCtx = ctx
					return &repositories.PaginatedPosts{}, nil
				},
			}
			listPostsController := listposts.NewListPostsController(listposts.NewListPostsControllerInput{ListPostsUseCase: listPostsUseCase})
			ginContext, _ := newListPostsRequest(test, "")
			ginContext.Request = ginContext.Request.WithContext(context.WithValue(ginContext.Request.Context(), contextKey{}, "request"))

			listPostsController.Handle(ginContext)

			if receivedCtx == nil || receivedCtx.Value(contextKey{}) != "request" {
				test.Errorf("Expected the use case to receive the request context, got %v", receivedCtx)
			}
		})

		test.Run("Should respond 200 with the page wrapped in the response envelope", func(test *testing.T) {
			createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
			nextCursor := "next-cursor"
			listPostsUseCase := &listPostsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input listposts.ListPostsInput) (*repositories.PaginatedPosts, error) {
					return &repositories.PaginatedPosts{
						Items:      []*entities.Post{{ID: "post-id", Title: "title", Content: "content", CreatedAt: createdAt, UpdatedAt: createdAt}},
						NextCursor: &nextCursor,
					}, nil
				},
			}
			listPostsController := listposts.NewListPostsController(listposts.NewListPostsControllerInput{ListPostsUseCase: listPostsUseCase})
			ginContext, recorder := newListPostsRequest(test, "")

			listPostsController.Handle(ginContext)

			if recorder.Code != http.StatusOK {
				test.Errorf("Status = %d, want %d", recorder.Code, http.StatusOK)
			}
			want := `{"data":{"items":[{"id":"post-id","title":"title","content":"content","createdAt":"2025-01-01T10:00:00Z","updatedAt":"2025-01-01T10:00:00Z"}],"nextCursor":"next-cursor"},"message":"Posts listed successfully"}`
			if recorder.Body.String() != want {
				test.Errorf("Body = %s, want %s", recorder.Body.String(), want)
			}
			if len(ginContext.Errors) != 0 {
				test.Errorf("Expected no errors registered, got %v", ginContext.Errors)
			}
		})

		test.Run("Should register the use case error and not write a response", func(test *testing.T) {
			useCaseError := errors.New("use case failed")
			listPostsUseCase := &listPostsUseCaseMock{
				ExecuteFunc: func(ctx context.Context, input listposts.ListPostsInput) (*repositories.PaginatedPosts, error) {
					return nil, useCaseError
				},
			}
			listPostsController := listposts.NewListPostsController(listposts.NewListPostsControllerInput{ListPostsUseCase: listPostsUseCase})
			ginContext, _ := newListPostsRequest(test, "")

			listPostsController.Handle(ginContext)

			lastError := ginContext.Errors.Last()
			if lastError == nil || !errors.Is(lastError.Err, useCaseError) {
				test.Errorf("Registered error = %v, want %v", lastError, useCaseError)
			}
			if ginContext.Writer.Written() {
				test.Error("Expected no response written, the error middleware is responsible for it")
			}
		})

		invalidQueries := []struct {
			name  string
			query string
		}{
			{name: "limit is not a number", query: "?limit=abc"},
			{name: "limit is zero", query: "?limit=0"},
			{name: "limit is negative", query: "?limit=-1"},
			{name: "limit is above the maximum", query: "?limit=101"},
		}
		for _, tc := range invalidQueries {
			test.Run("Should register an InvalidRequestParamsError and not call the use case when "+tc.name, func(test *testing.T) {
				listPostsUseCase := &listPostsUseCaseMock{}
				listPostsController := listposts.NewListPostsController(listposts.NewListPostsControllerInput{ListPostsUseCase: listPostsUseCase})
				ginContext, _ := newListPostsRequest(test, tc.query)

				listPostsController.Handle(ginContext)

				if listPostsUseCase.calls != 0 {
					test.Errorf("Expected the use case not to be called, got %d calls", listPostsUseCase.calls)
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
