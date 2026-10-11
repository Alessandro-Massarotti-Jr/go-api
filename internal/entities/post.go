package entities

import (
	"time"

	"github.com/google/uuid"
)

type PostId string

type Post struct {
	ID        PostId    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreatePostInput holds the data used to create a Post.
// Pointer fields are optional: when nil, a default value is applied.
type CreatePostInput struct {
	ID        *PostId // default: new UUID
	Title     string
	Content   string
	CreatedAt *time.Time // default: time.Now()
	UpdatedAt *time.Time // default: time.Now()
}

func Create(input CreatePostInput) *Post {

	var id PostId
	var createdAt, updatedAt time.Time
	if input.ID != nil {
		id = *input.ID
	} else {
		id = PostId(uuid.New().String())
	}

	now := time.Now()

	if input.CreatedAt != nil {
		createdAt = *input.CreatedAt
	} else {
		createdAt = now
	}

	if input.UpdatedAt != nil {
		updatedAt = *input.UpdatedAt
	} else {
		updatedAt = now
	}

	return &Post{
		ID:        id,
		Title:     input.Title,
		Content:   input.Content,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
