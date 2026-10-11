package entities

import (
	"time"

	"github.com/google/uuid"
)

type PostCommentId string

type PostComment struct {
	ID        PostCommentId `json:"id"`
	PostID    PostId        `json:"postId"`
	Comment   string        `json:"comment"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

// CreatePostCommentInput holds the data used to create a PostComment.
// Pointer fields are optional: when nil, a default value is applied.
type CreatePostCommentInput struct {
	ID        *PostCommentId // default: new UUID
	PostID    PostId
	Comment   string
	CreatedAt *time.Time // default: time.Now()
	UpdatedAt *time.Time // default: time.Now()
}

func CreatePostComment(input CreatePostCommentInput) *PostComment {

	var id PostCommentId
	var createdAt, updatedAt time.Time
	if input.ID != nil {
		id = *input.ID
	} else {
		id = PostCommentId(uuid.New().String())
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

	return &PostComment{
		ID:        id,
		PostID:    input.PostID,
		Comment:   input.Comment,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
