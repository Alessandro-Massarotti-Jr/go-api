package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PostSchema struct {
	MongoID   bson.ObjectID `bson:"_id,omitempty"`
	ID        string        `bson:"id"`
	Title     string        `bson:"title"`
	Content   string        `bson:"content"`
	CreatedAt time.Time     `bson:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt"`
}

var PostIndexes = []mongo.IndexModel{
	{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
	{Keys: bson.D{{Key: "createdAt", Value: -1}, {Key: "id", Value: -1}}},
}
