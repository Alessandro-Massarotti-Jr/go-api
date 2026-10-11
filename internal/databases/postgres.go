package databases

import (
	"context"
	"sync"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/config"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

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
	appLogger := logger.GetInstance()

	client, err := gorm.Open(postgres.Open(config.GetConfig(config.PostgresDSN)), &gorm.Config{
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

	appLogger.Info(logger.LogMessage{
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
