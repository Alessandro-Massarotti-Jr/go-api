package repositories

import (
	"context"
	"sync"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/databases"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/schemas"
	"gorm.io/gorm"
)

const postCommentTable = "post_comments"

type GormPostCommentRepository struct {
	postgres *databases.Postgres
}

var _ PostCommentRepository = (*GormPostCommentRepository)(nil)

var (
	gormPostCommentRepositoryInstance *GormPostCommentRepository
	gormPostCommentRepositoryErr      error
	gormPostCommentRepositoryOnce     sync.Once
)

// GetGormPostCommentRepositoryInstance returns the shared repository, building it on the first call
// with the shared Postgres instance and migrating the post_comments table (columns + indexes).
func GetGormPostCommentRepositoryInstance() (*GormPostCommentRepository, error) {
	gormPostCommentRepositoryOnce.Do(func() {
		gormPostCommentRepositoryInstance, gormPostCommentRepositoryErr = newGormPostCommentRepository()
	})
	return gormPostCommentRepositoryInstance, gormPostCommentRepositoryErr
}

func newGormPostCommentRepository() (*GormPostCommentRepository, error) {
	const location = "GormPostCommentRepository.New"

	postgres, err := databases.GetPostgresInstance()
	if err != nil {
		return nil, appErrors.CreatePostCommentRepositoryError(appErrors.CreatePostCommentRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to connect to Postgres",
			OriginError: err,
		})
	}

	gormPostCommentRepository := &GormPostCommentRepository{postgres: postgres}

	if err := gormPostCommentRepository.postCommentTable().AutoMigrate(&schemas.PostCommentSchema{}); err != nil {
		return nil, appErrors.CreatePostCommentRepositoryError(appErrors.CreatePostCommentRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to migrate post_comments table",
			OriginError: err,
		})
	}

	return gormPostCommentRepository, nil
}

func (gormPostCommentRepository *GormPostCommentRepository) postCommentTable() *gorm.DB {
	return gormPostCommentRepository.postgres.Table(databases.TableInput{Name: postCommentTable})
}

func (gormPostCommentRepository *GormPostCommentRepository) FindAllByPostId(ctx context.Context, input FindPostCommentsByPostIdInput) (*PaginatedPostComments, error) {
	const location = "GormPostCommentRepository.FindAllByPostId"

	limit := input.Limit
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}

	query := gormPostCommentRepository.postCommentTable().WithContext(ctx).Where("post_id = ?", string(input.PostID))
	if input.Cursor != nil && *input.Cursor != "" {
		cursor, err := decodeCursor(*input.Cursor)
		if err != nil {
			return nil, appErrors.CreateInvalidCursorError(appErrors.CreateInvalidCursorErrorInput{
				Location:    location,
				OriginError: err,
			})
		}
		query = query.Where("(created_at, id) < (?, ?)", cursor.CreatedAt, cursor.ID)
	}

	// Fetch one extra row to know whether there is a next page.
	var documents []schemas.PostCommentSchema
	err := query.
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit + 1).
		Find(&documents).Error
	if err != nil {
		return nil, appErrors.CreatePostCommentRepositoryError(appErrors.CreatePostCommentRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to find post comments",
			OriginError: err,
		})
	}

	var nextCursor *string
	if len(documents) > limit {
		documents = documents[:limit]
		last := documents[len(documents)-1]
		encoded, err := encodeCursor(pageCursor{CreatedAt: last.CreatedAt, ID: last.ID})
		if err != nil {
			return nil, appErrors.CreatePostCommentRepositoryError(appErrors.CreatePostCommentRepositoryErrorInput{
				Location:    location,
				Message:     "Failed to encode pagination cursor",
				OriginError: err,
			})
		}
		nextCursor = &encoded
	}

	postComments := make([]*entities.PostComment, 0, len(documents))
	for _, document := range documents {
		postComments = append(postComments, toPostCommentEntity(document))
	}

	return &PaginatedPostComments{Items: postComments, NextCursor: nextCursor}, nil
}

func (gormPostCommentRepository *GormPostCommentRepository) Create(ctx context.Context, postComment *entities.PostComment) (*entities.PostComment, error) {
	const location = "GormPostCommentRepository.Create"

	document := toPostCommentSchema(postComment)
	if err := gormPostCommentRepository.postCommentTable().WithContext(ctx).Create(&document).Error; err != nil {
		return nil, appErrors.CreatePostCommentRepositoryError(appErrors.CreatePostCommentRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to create post comment",
			OriginError: err,
		})
	}

	return toPostCommentEntity(document), nil
}

func (gormPostCommentRepository *GormPostCommentRepository) DeleteAllByPostId(ctx context.Context, input DeletePostCommentsByPostIdInput) error {
	const location = "GormPostCommentRepository.DeleteAllByPostId"

	err := gormPostCommentRepository.postCommentTable().WithContext(ctx).
		Where("post_id = ?", string(input.PostID)).
		Delete(&schemas.PostCommentSchema{}).Error
	if err != nil {
		return appErrors.CreatePostCommentRepositoryError(appErrors.CreatePostCommentRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to delete post comments",
			OriginError: err,
		})
	}

	return nil
}

// toPostCommentSchema truncates dates to microseconds, the precision Postgres stores,
// so the entity built back from the schema matches what is persisted.
func toPostCommentSchema(postComment *entities.PostComment) schemas.PostCommentSchema {
	return schemas.PostCommentSchema{
		ID:        string(postComment.ID),
		PostID:    string(postComment.PostID),
		Comment:   postComment.Comment,
		CreatedAt: postComment.CreatedAt.UTC().Truncate(time.Microsecond),
		UpdatedAt: postComment.UpdatedAt.UTC().Truncate(time.Microsecond),
	}
}

func toPostCommentEntity(schema schemas.PostCommentSchema) *entities.PostComment {
	id := entities.PostCommentId(schema.ID)
	createdAt := schema.CreatedAt.UTC()
	updatedAt := schema.UpdatedAt.UTC()

	return entities.CreatePostComment(entities.CreatePostCommentInput{
		ID:        &id,
		PostID:    entities.PostId(schema.PostID),
		Comment:   schema.Comment,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	})
}
