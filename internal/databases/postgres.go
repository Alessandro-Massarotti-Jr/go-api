package databases

import (
	"context"
	"os"
	"sync"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

const defaultPostgresDSN = "host=localhost user=postgres password=postgres dbname=go_api port=5432 sslmode=disable"

type Postgres struct {
	client *gorm.DB
}

var (
	postgresInstance    *Postgres
	postgresInstanceErr error
	postgresOnce        sync.Once
)

func GetPostgresInstance() (*Postgres, error) {
	postgresOnce.Do(func() {
		postgresInstance, postgresInstanceErr = connectPostgres()
	})
	return postgresInstance, postgresInstanceErr
}

func connectPostgres() (*Postgres, error) {
	logger := config.GetInstance()

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = defaultPostgresDSN
	}

	client, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         gormLogger.Default.LogMode(gormLogger.Silent),
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := client.DB()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	logger.Info(config.LogMessage{
		Action:  "postgres.connect",
		Message: "Connected to Postgres",
	})

	return &Postgres{client: client}, nil
}

type TableInput struct {
	Name string
}

// Table returns a gorm session scoped to the given table.
func (p *Postgres) Table(input TableInput) *gorm.DB {
	return p.client.Table(input.Name)
}

func (p *Postgres) Disconnect() error {
	sqlDB, err := p.client.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
