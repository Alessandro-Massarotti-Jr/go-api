package repositories

import (
	"context"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
)

const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)

// FindPostsInput holds the cursor pagination params.
// Cursor is the NextCursor returned by the previous page; nil fetches the first page.
type FindPostsInput struct {
	Cursor *string
	Limit  int // default: DefaultPageLimit, capped at MaxPageLimit
}

type PaginatedPosts struct {
	Items      []*entities.Post `json:"items"`
	NextCursor *string          `json:"nextCursor"`
}

type FindPostByIdInput struct {
	ID entities.PostId
}

type UpdatePostInput struct {
	ID      entities.PostId
	Title   *string
	Content *string
}

type DeletePostInput struct {
	ID entities.PostId
}

type PostRepository interface {
	FindAll(ctx context.Context, input FindPostsInput) (*PaginatedPosts, error)
	FindById(ctx context.Context, input FindPostByIdInput) (*entities.Post, error)
	Create(ctx context.Context, post *entities.Post) (*entities.Post, error)
	Update(ctx context.Context, input UpdatePostInput) (*entities.Post, error)
	Delete(ctx context.Context, input DeletePostInput) error
}
