package repositories

import (
	"context"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
)

// FindPostCommentsByPostIdInput holds the post whose comments are listed and the cursor pagination params.
// Cursor is the NextCursor returned by the previous page; nil fetches the first page.
type FindPostCommentsByPostIdInput struct {
	PostID entities.PostId
	Cursor *string
	Limit  int // default: DefaultPageLimit, capped at MaxPageLimit
}

type PaginatedPostComments struct {
	Items      []*entities.PostComment `json:"items"`
	NextCursor *string                 `json:"nextCursor"`
}

type PostCommentRepository interface {
	FindAllByPostId(ctx context.Context, input FindPostCommentsByPostIdInput) (*PaginatedPostComments, error)
	Create(ctx context.Context, postComment *entities.PostComment) (*entities.PostComment, error)
}
