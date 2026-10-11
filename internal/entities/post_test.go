package entities_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	"github.com/google/uuid"
)

// createWithinWindow calls fn and returns its result along with the time window around the call.
func createWithinWindow(fn func() *entities.Post) (post *entities.Post, before time.Time, after time.Time) {
	before = time.Now()
	post = fn()
	after = time.Now()
	return post, before, after
}

func isWithin(value, before, after time.Time) bool {
	return !value.Before(before) && !value.After(after)
}

func TestPost(test *testing.T) {

	test.Run("Create", func(test *testing.T) {
		test.Run("Should keep the title and content", func(test *testing.T) {
			post := entities.Create(entities.CreatePostInput{Title: "My title", Content: "My content"})

			if post.Title != "My title" {
				test.Errorf("Title = %q, want %q", post.Title, "My title")
			}
			if post.Content != "My content" {
				test.Errorf("Content = %q, want %q", post.Content, "My content")
			}
		})

		test.Run("Should accept empty title and content", func(test *testing.T) {
			post := entities.Create(entities.CreatePostInput{})

			if post.Title != "" || post.Content != "" {
				test.Errorf("Expected empty title and content, got %q and %q", post.Title, post.Content)
			}
		})

		test.Run("Should generate a valid UUID when not provided", func(test *testing.T) {
			post := entities.Create(entities.CreatePostInput{Title: "title"})

			if _, err := uuid.Parse(string(post.ID)); err != nil {
				test.Errorf("ID %q is not a valid UUID: %v", post.ID, err)
			}
		})

		test.Run("Should generate a different ID for each post", func(test *testing.T) {
			post1 := entities.Create(entities.CreatePostInput{Title: "title"})
			post2 := entities.Create(entities.CreatePostInput{Title: "title"})

			if post1.ID == post2.ID {
				test.Errorf("Expected different IDs, got %q twice", post1.ID)
			}
		})

		test.Run("Should use the provided ID", func(test *testing.T) {
			id := entities.PostId("custom-id")
			post := entities.Create(entities.CreatePostInput{ID: &id})

			if post.ID != id {
				test.Errorf("ID = %q, want %q", post.ID, id)
			}
		})

		test.Run("Should not be affected by changes to the input after creation", func(test *testing.T) {
			id := entities.PostId("custom-id")
			post := entities.Create(entities.CreatePostInput{ID: &id})
			id = "changed"

			if post.ID != "custom-id" {
				test.Errorf("ID = %q, want %q", post.ID, "custom-id")
			}
		})

		test.Run("Should default to now when not provided", func(test *testing.T) {
			post, before, after := createWithinWindow(func() *entities.Post {
				return entities.Create(entities.CreatePostInput{Title: "title"})
			})

			if !isWithin(post.CreatedAt, before, after) {
				test.Errorf("CreatedAt = %v, want between %v and %v", post.CreatedAt, before, after)
			}
		})

		test.Run("Should use the provided date", func(test *testing.T) {
			createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
			post := entities.Create(entities.CreatePostInput{CreatedAt: &createdAt})

			if !post.CreatedAt.Equal(createdAt) {
				test.Errorf("CreatedAt = %v, want %v", post.CreatedAt, createdAt)
			}
		})

		test.Run("Should not be affected by changes to the input after creation", func(test *testing.T) {
			createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
			post := entities.Create(entities.CreatePostInput{CreatedAt: &createdAt})
			createdAt = createdAt.Add(time.Hour)

			want := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
			if !post.CreatedAt.Equal(want) {
				test.Errorf("CreatedAt = %v, want %v", post.CreatedAt, want)
			}
		})

		test.Run("Should default to now when not provided", func(test *testing.T) {
			post, before, after := createWithinWindow(func() *entities.Post {
				return entities.Create(entities.CreatePostInput{Title: "title"})
			})

			if !isWithin(post.UpdatedAt, before, after) {
				test.Errorf("UpdatedAt = %v, want between %v and %v", post.UpdatedAt, before, after)
			}
		})

		test.Run("Should be equal to CreatedAt when both are not provided", func(test *testing.T) {
			post := entities.Create(entities.CreatePostInput{Title: "title"})

			if !post.UpdatedAt.Equal(post.CreatedAt) {
				test.Errorf("UpdatedAt = %v, want equal to CreatedAt %v", post.UpdatedAt, post.CreatedAt)
			}
		})

		test.Run("Should default to now even when CreatedAt is provided", func(test *testing.T) {
			createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
			post, before, after := createWithinWindow(func() *entities.Post {
				return entities.Create(entities.CreatePostInput{CreatedAt: &createdAt})
			})

			if !isWithin(post.UpdatedAt, before, after) {
				test.Errorf("UpdatedAt = %v, want between %v and %v", post.UpdatedAt, before, after)
			}
		})

		test.Run("Should use the provided date", func(test *testing.T) {
			updatedAt := time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)
			post := entities.Create(entities.CreatePostInput{UpdatedAt: &updatedAt})

			if !post.UpdatedAt.Equal(updatedAt) {
				test.Errorf("UpdatedAt = %v, want %v", post.UpdatedAt, updatedAt)
			}
		})

		test.Run("Should keep every provided value", func(test *testing.T) {
			id := entities.PostId("custom-id")
			createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
			updatedAt := time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)

			post := entities.Create(entities.CreatePostInput{
				ID:        &id,
				Title:     "title",
				Content:   "content",
				CreatedAt: &createdAt,
				UpdatedAt: &updatedAt,
			})

			want := entities.Post{
				ID:        id,
				Title:     "title",
				Content:   "content",
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}
			if *post != want {
				test.Errorf("Post = %+v, want %+v", *post, want)
			}
		})
	})

	test.Run("Should serialize with camelCase keys and RFC3339 dates", func(test *testing.T) {
		id := entities.PostId("custom-id")
		createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
		updatedAt := time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)

		post := entities.Create(entities.CreatePostInput{
			ID:        &id,
			Title:     "title",
			Content:   "content",
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		})

		output, err := json.Marshal(post)
		if err != nil {
			test.Fatalf("Failed to marshal post: %v", err)
		}

		want := `{"id":"custom-id","title":"title","content":"content","totalComments":0,"createdAt":"2025-01-01T10:00:00Z","updatedAt":"2025-01-02T10:00:00Z"}`
		if string(output) != want {
			test.Errorf("JSON = %s, want %s", output, want)
		}
	})

}
