package events

import "time"

// PostCommentCreatedTopic is published whenever a comment is added to a post.
const PostCommentCreatedTopic = "postCommentCreated"

type PostCommentCreatedEvent struct {
	ID        string    `json:"id"`
	PostID    string    `json:"postId"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}
