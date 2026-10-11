package schemas

import "time"

// PostCommentSchema is the Postgres row of a post comment. Indexes are declared in the gorm tags:
// idx_post_comments_post_id_created_at_id backs the comments listing (filter by post, sort by createdAt desc + id desc).
type PostCommentSchema struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey;index:idx_post_comments_post_id_created_at_id,priority:3,sort:desc"`
	PostID    string    `gorm:"column:post_id;type:uuid;not null;index:idx_post_comments_post_id_created_at_id,priority:1"`
	Comment   string    `gorm:"column:comment;type:text;not null"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;index:idx_post_comments_post_id_created_at_id,priority:2,sort:desc"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null"`
}
