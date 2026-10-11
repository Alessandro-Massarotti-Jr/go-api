package config

import (
	"errors"
	"io/fs"
	"os"
	"sync"

	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/joho/godotenv"
)

// Key is the name of a config value, used like an enum: GetConfig(config.Port).
type Key string

const (
	Port    Key = "PORT"
	AppName Key = "APP_NAME"

	MongoURI             Key = "MONGO_URI"
	MongoDatabase        Key = "MONGO_DATABASE"
	MongoPostsCollection Key = "MONGO_POSTS_COLLECTION"

	PostgresDSN Key = "POSTGRES_DSN"

	RabbitMQURI Key = "RABBITMQ_URI"
)

type keyDefinition struct {
	key          Key
	defaultValue string
	required     bool
}

// definitions lists every key loaded from the environment.
// A value falls back to defaultValue when unset/empty; a required one without value fails loading.
var definitions = []keyDefinition{
	{key: Port, defaultValue: "8080"},
	{key: AppName, defaultValue: "go-api"},

	{key: MongoURI, required: true},
	{key: MongoDatabase, defaultValue: "go_api"},
	{key: MongoPostsCollection, defaultValue: "posts"},

	{key: PostgresDSN, required: true},

	{key: RabbitMQURI, required: true},
}

const envFile = ".env"

type Config struct {
	values map[Key]string
}

var (
	configInstance *Config
	configErr      error
	configOnce     sync.Once
)

// GetInstance loads the .env file and the process env on the first call, applies defaults and
// returns an error when a required value is missing. Call it at the start of main.
func GetInstance() (*Config, error) {
	configOnce.Do(func() {
		values, err := loadValues()
		if err != nil {
			configErr = err
			return
		}
		configInstance = &Config{values: values}
	})
	return configInstance, configErr
}

// GetConfig returns the configured value for key. It panics when the config is invalid or the
// key is not in definitions, since both mean the process cannot run; main surfaces the load error
// earlier through GetInstance.
func GetConfig(key Key) string {
	instance, err := GetInstance()
	if err != nil {
		panic(err)
	}
	value, ok := instance.values[key]
	if !ok {
		panic("config.GetConfig: undefined key " + string(key))
	}
	return value
}

func loadValues() (map[Key]string, error) {
	const location = "config.load"

	// Vars already set in the process take precedence over the .env file.
	if err := godotenv.Load(envFile); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	values := make(map[Key]string, len(definitions))
	var missingKeys []string
	for _, definition := range definitions {
		value := os.Getenv(string(definition.key))
		if value == "" {
			value = definition.defaultValue
		}
		if value == "" && definition.required {
			missingKeys = append(missingKeys, string(definition.key))
			continue
		}
		values[definition.key] = value
	}

	if len(missingKeys) > 0 {
		return nil, appErrors.CreateMissingEnvError(appErrors.CreateMissingEnvErrorInput{
			Location: location,
			Keys:     missingKeys,
		})
	}

	return values, nil
}
