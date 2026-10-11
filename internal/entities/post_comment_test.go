package entities_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/google/uuid"
)

func TestPostComment(test *testing.T) {

	test.Run("CreatePostComment", func(test *testing.T) {
		test.Run("Should keep the post ID and comment", func(test *testing.T) {
			postComment := entities.CreatePostComment(entities.CreatePostCommentInput{PostID: "post-id", Comment: "My comment"})

			if postComment.PostID != "post-id" {
				test.Errorf("PostID = %q, want %q", postComment.PostID, "post-id")
			}
			if postComment.Comment != "My comment" {
				test.Errorf("Comment = %q, want %q", postComment.Comment, "My comment")
			}
		})

		test.Run("Should generate a valid UUID when not provided", func(test *testing.T) {
			postComment := entities.CreatePostComment(entities.CreatePostCommentInput{PostID: "post-id"})

			if _, err := uuid.Parse(string(postComment.ID)); err != nil {
				test.Errorf("ID %q is not a valid UUID: %v", postComment.ID, err)
			}
		})

		test.Run("Should use the provided ID", func(test *testing.T) {
			id := entities.PostCommentId("custom-id")
			postComment := entities.CreatePostComment(entities.CreatePostCommentInput{ID: &id})

			if postComment.ID != id {
				test.Errorf("ID = %q, want %q", postComment.ID, id)
			}
		})

		test.Run("Should default dates to now when not provided", func(test *testing.T) {
			before := time.Now()
			postComment := entities.CreatePostComment(entities.CreatePostCommentInput{PostID: "post-id"})
			after := time.Now()

			if !isWithin(postComment.CreatedAt, before, after) {
				test.Errorf("CreatedAt = %v, want between %v and %v", postComment.CreatedAt, before, after)
			}
			if !postComment.UpdatedAt.Equal(postComment.CreatedAt) {
				test.Errorf("UpdatedAt = %v, want equal to CreatedAt %v", postComment.UpdatedAt, postComment.CreatedAt)
			}
		})

		test.Run("Should use the provided dates", func(test *testing.T) {
			createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
			updatedAt := time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)
			postComment := entities.CreatePostComment(entities.CreatePostCommentInput{CreatedAt: &createdAt, UpdatedAt: &updatedAt})

			if !postComment.CreatedAt.Equal(createdAt) {
				test.Errorf("CreatedAt = %v, want %v", postComment.CreatedAt, createdAt)
			}
			if !postComment.UpdatedAt.Equal(updatedAt) {
				test.Errorf("UpdatedAt = %v, want %v", postComment.UpdatedAt, updatedAt)
			}
		})
	})

	test.Run("Should serialize with camelCase keys and RFC3339 dates", func(test *testing.T) {
		id := entities.PostCommentId("custom-id")
		createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
		updatedAt := time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)
		postComment := entities.CreatePostComment(entities.CreatePostCommentInput{
			ID:        &id,
			PostID:    "post-id",
			Comment:   "comment",
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		})

		output, err := json.Marshal(postComment)
		if err != nil {
			test.Fatalf("Failed to marshal post comment: %v", err)
		}

		want := `{"id":"custom-id","postId":"post-id","comment":"comment","createdAt":"2025-01-01T10:00:00Z","updatedAt":"2025-01-02T10:00:00Z"}`
		if string(output) != want {
			test.Errorf("JSON = %s, want %s", output, want)
		}
	})

}
