package events

import "time"

// PostDeletedTopic is published whenever a post is deleted.
const PostDeletedTopic = "postDeleted"

type PostDeletedEvent struct {
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}
