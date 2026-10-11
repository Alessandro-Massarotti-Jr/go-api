package repositories

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/databases"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/entities"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/schemas"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoPostRepository struct {
	postCollection *mongo.Collection
}

var _ PostRepository = (*MongoPostRepository)(nil)

const createIndexesTimeout = 10 * time.Second

var (
	mongoPostRepositoryInstance *MongoPostRepository
	mongoPostRepositoryErr      error
	mongoPostRepositoryOnce     sync.Once
)

// GetMongoPostRepositoryInstance returns the shared repository, building it on the first call
// with the shared Mongo instance and ensuring the post indexes exist.
func GetMongoPostRepositoryInstance() (*MongoPostRepository, error) {
	mongoPostRepositoryOnce.Do(func() {
		mongoPostRepositoryInstance, mongoPostRepositoryErr = newMongoPostRepository()
	})
	return mongoPostRepositoryInstance, mongoPostRepositoryErr
}

func newMongoPostRepository() (*MongoPostRepository, error) {
	const location = "MongoPostRepository.New"

	db, err := databases.GetInstance()
	if err != nil {
		return nil, appErrors.CreatePostRepositoryError(appErrors.CreatePostRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to connect to MongoDB",
			OriginError: err,
		})
	}

	postCollection := db.Collection(databases.CollectionInput{Name: "posts"})

	ctx, cancel := context.WithTimeout(context.Background(), createIndexesTimeout)
	defer cancel()
	if _, err := postCollection.Indexes().CreateMany(ctx, schemas.PostIndexes); err != nil {
		return nil, appErrors.CreatePostRepositoryError(appErrors.CreatePostRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to create posts indexes",
			OriginError: err,
		})
	}

	return &MongoPostRepository{postCollection: postCollection}, nil
}

func (mongoPostRepository *MongoPostRepository) FindAll(ctx context.Context, input FindPostsInput) (*PaginatedPosts, error) {
	const location = "MongoPostRepository.FindAll"

	limit := input.Limit
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}

	filter := bson.M{}
	if input.Cursor != nil && *input.Cursor != "" {
		cursor, err := decodeCursor(*input.Cursor)
		if err != nil {
			return nil, appErrors.CreateInvalidCursorError(appErrors.CreateInvalidCursorErrorInput{
				Location:    location,
				OriginError: err,
			})
		}
		filter = bson.M{"$or": bson.A{
			bson.M{"createdAt": bson.M{"$lt": cursor.CreatedAt}},
			bson.M{"createdAt": cursor.CreatedAt, "id": bson.M{"$lt": cursor.ID}},
		}}
	}

	// Fetch one extra document to know whether there is a next page.
	findOptions := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "id", Value: -1}}).
		SetLimit(int64(limit + 1))

	result, err := mongoPostRepository.postCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, appErrors.CreatePostRepositoryError(appErrors.CreatePostRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to find posts",
			OriginError: err,
		})
	}

	var documents []schemas.PostSchema
	if err := result.All(ctx, &documents); err != nil {
		return nil, appErrors.CreatePostRepositoryError(appErrors.CreatePostRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to decode posts",
			OriginError: err,
		})
	}

	var nextCursor *string
	if len(documents) > limit {
		documents = documents[:limit]
		last := documents[len(documents)-1]
		encoded, err := encodeCursor(pageCursor{CreatedAt: last.CreatedAt, ID: last.ID})
		if err != nil {
			return nil, appErrors.CreatePostRepositoryError(appErrors.CreatePostRepositoryErrorInput{
				Location:    location,
				Message:     "Failed to encode pagination cursor",
				OriginError: err,
			})
		}
		nextCursor = &encoded
	}

	posts := make([]*entities.Post, 0, len(documents))
	for _, document := range documents {
		posts = append(posts, toEntity(document))
	}

	return &PaginatedPosts{Items: posts, NextCursor: nextCursor}, nil
}

func (mongoPostRepository *MongoPostRepository) FindById(ctx context.Context, input FindPostByIdInput) (*entities.Post, error) {
	const location = "MongoPostRepository.FindById"

	var document schemas.PostSchema
	err := mongoPostRepository.postCollection.FindOne(ctx, bson.M{"id": string(input.ID)}).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, appErrors.CreatePostNotFoundError(appErrors.CreatePostNotFoundErrorInput{
			Location: location,
			ID:       string(input.ID),
		})
	}
	if err != nil {
		return nil, appErrors.CreatePostRepositoryError(appErrors.CreatePostRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to find post",
			OriginError: err,
		})
	}

	return toEntity(document), nil
}

func (mongoPostRepository *MongoPostRepository) Create(ctx context.Context, post *entities.Post) (*entities.Post, error) {
	const location = "MongoPostRepository.Create"

	document := toSchema(post)
	if _, err := mongoPostRepository.postCollection.InsertOne(ctx, document); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, appErrors.CreatePostAlreadyExistsError(appErrors.CreatePostAlreadyExistsErrorInput{
				Location:    location,
				ID:          string(post.ID),
				OriginError: err,
			})
		}
		return nil, appErrors.CreatePostRepositoryError(appErrors.CreatePostRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to create post",
			OriginError: err,
		})
	}

	return toEntity(document), nil
}

func (mongoPostRepository *MongoPostRepository) Update(ctx context.Context, input UpdatePostInput) (*entities.Post, error) {
	const location = "MongoPostRepository.Update"

	set := bson.M{"updatedAt": time.Now().UTC().Truncate(time.Millisecond)}
	if input.Title != nil {
		set["title"] = *input.Title
	}
	if input.Content != nil {
		set["content"] = *input.Content
	}

	var document schemas.PostSchema
	err := mongoPostRepository.postCollection.FindOneAndUpdate(
		ctx,
		bson.M{"id": string(input.ID)},
		bson.M{"$set": set},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, appErrors.CreatePostNotFoundError(appErrors.CreatePostNotFoundErrorInput{
			Location: location,
			ID:       string(input.ID),
		})
	}
	if err != nil {
		return nil, appErrors.CreatePostRepositoryError(appErrors.CreatePostRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to update post",
			OriginError: err,
		})
	}

	return toEntity(document), nil
}

func (mongoPostRepository *MongoPostRepository) Delete(ctx context.Context, input DeletePostInput) error {
	const location = "MongoPostRepository.Delete"

	result, err := mongoPostRepository.postCollection.DeleteOne(ctx, bson.M{"id": string(input.ID)})
	if err != nil {
		return appErrors.CreatePostRepositoryError(appErrors.CreatePostRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to delete post",
			OriginError: err,
		})
	}
	if result.DeletedCount == 0 {
		return appErrors.CreatePostNotFoundError(appErrors.CreatePostNotFoundErrorInput{
			Location: location,
			ID:       string(input.ID),
		})
	}

	return nil
}

func (mongoPostRepository *MongoPostRepository) IncrementTotalComments(ctx context.Context, input IncrementPostTotalCommentsInput) error {
	const location = "MongoPostRepository.IncrementTotalComments"

	result, err := mongoPostRepository.postCollection.UpdateOne(
		ctx,
		bson.M{"id": string(input.ID)},
		bson.M{"$inc": bson.M{"totalComments": 1}},
	)
	if err != nil {
		return appErrors.CreatePostRepositoryError(appErrors.CreatePostRepositoryErrorInput{
			Location:    location,
			Message:     "Failed to increment post total comments",
			OriginError: err,
		})
	}
	if result.MatchedCount == 0 {
		return appErrors.CreatePostNotFoundError(appErrors.CreatePostNotFoundErrorInput{
			Location: location,
			ID:       string(input.ID),
		})
	}

	return nil
}

// toSchema truncates dates to milliseconds, the precision Mongo stores,
// so the entity built back from the schema matches what is persisted.
func toSchema(post *entities.Post) schemas.PostSchema {
	return schemas.PostSchema{
		ID:            string(post.ID),
		Title:         post.Title,
		Content:       post.Content,
		TotalComments: post.TotalComments,
		CreatedAt:     post.CreatedAt.UTC().Truncate(time.Millisecond),
		UpdatedAt:     post.UpdatedAt.UTC().Truncate(time.Millisecond),
	}
}

func toEntity(schema schemas.PostSchema) *entities.Post {
	id := entities.PostId(schema.ID)
	createdAt := schema.CreatedAt.UTC()
	updatedAt := schema.UpdatedAt.UTC()

	return entities.Create(entities.CreatePostInput{
		ID:            &id,
		Title:         schema.Title,
		Content:       schema.Content,
		TotalComments: schema.TotalComments,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	})
}
