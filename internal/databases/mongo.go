package databases

import (
	"context"
	"sync"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/config"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/logger"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const connectTimeout = 10 * time.Second

type Mongo struct {
	client   *mongo.Client
	database *mongo.Database
}

var (
	instance    *Mongo
	instanceErr error
	once        sync.Once
)

func GetInstance() (*Mongo, error) {
	once.Do(func() {
		instance, instanceErr = connect()
	})
	return instance, instanceErr
}

func connect() (*Mongo, error) {
	appLogger := logger.GetInstance()
	databaseName := config.GetConfig(config.MongoDatabase)

	client, err := mongo.Connect(options.Client().ApplyURI(config.GetConfig(config.MongoURI)))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	appLogger.Info(logger.LogMessage{
		Action:  "mongo.connect",
		Message: "Connected to MongoDB",
		Data:    map[string]any{"database": databaseName},
	})

	return &Mongo{
		client:   client,
		database: client.Database(databaseName),
	}, nil
}

type CollectionInput struct {
	Name string
}

func (m *Mongo) Collection(input CollectionInput) *mongo.Collection {
	return m.database.Collection(input.Name)
}

func (m *Mongo) Disconnect(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
