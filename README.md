# Go API

<p>
  <img src="https://img.shields.io/badge/made%20by-Alessandro%20Massarotti%20Jr-00ADD8?style=flat-square">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square">
  <img alt="GitHub language count" src="https://img.shields.io/github/languages/count/Alessandro-Massarotti-Jr/go-api?color=00ADD8&style=flat-square">
  <img alt="GitHub Top Language" src="https://img.shields.io/github/languages/top/Alessandro-Massarotti-Jr/go-api?color=00ADD8&style=flat-square">
</p>

A study project on how to structure a Go application: a small posts & comments API organized around use cases, with an HTTP server and an event listener communicating through RabbitMQ.

## Tech stack

- [Gin](https://github.com/gin-gonic/gin) — HTTP server
- [MongoDB](https://www.mongodb.com/) — posts storage
- [PostgreSQL](https://www.postgresql.org/) + [GORM](https://gorm.io/) — comments storage
- [RabbitMQ](https://www.rabbitmq.com/) — domain events between services
- [Swag](https://github.com/swaggo/swag) — Swagger / OpenAPI docs

## Project structure

```
cmd/
  http/          # HTTP server entrypoint
  listener/      # RabbitMQ event listener entrypoint
docs/            # Generated Swagger files
internal/
  config/        # Environment config (validated on startup)
  databases/     # Mongo and Postgres connections
  entities/      # Domain entities
  errors/        # Named application errors
  events/        # Domain events
  listener/      # Event listener registration
  logger/        # Logger
  middlewares/   # HTTP middlewares
  providers/     # Event provider (RabbitMQ implementation)
  repositories/  # Persistence layer
  schemas/       # Database schemas
  server/        # Gin server and routes
  usecases/      # One folder per use case (use case, controller, wiring)
```

## Endpoints

| Method | Path                   | Description           |
| ------ | ---------------------- | --------------------- |
| GET    | `/healthy`             | Health check          |
| GET    | `/swagger/*any`        | Swagger UI            |
| POST   | `/posts`               | Create a post         |
| GET    | `/posts`               | List posts            |
| GET    | `/posts/:id`           | Get a post            |
| PATCH  | `/posts/:id`           | Update a post         |
| DELETE | `/posts/:id`           | Delete a post         |
| POST   | `/posts/:id/comments`  | Create a comment      |
| GET    | `/posts/:id/comments`  | List a post's comments |

## Events

The listener consumes domain events published by the HTTP service:

- **Post comment created** → increments the post's total comments
- **Post deleted** → deletes the post's comments

## Getting started

### With Docker Compose

Starts the HTTP server, the listener, MongoDB, PostgreSQL and RabbitMQ:

```bash
docker compose up --build
```

The API will be available at `http://localhost:8080` and Swagger at `http://localhost:8080/swagger/index.html`.

### Locally

Requirements: Go 1.26+ and running MongoDB, PostgreSQL and RabbitMQ instances.

```bash
cp .env.sample .env   # adjust values if needed
make run-http         # start the HTTP server
make run-listener     # start the event listener
```

### Other commands

```bash
make test      # run unit tests with coverage
make swagger   # regenerate Swagger docs
```

## Environment variables

See [.env.sample](.env.sample).

| Variable                 | Description                                                                 |
| ------------------------ | --------------------------------------------------------------------------- |
| `PORT`                   | HTTP server port                                                            |
| `APP_NAME`               | Service name, used in routing keys (`services.<APP_NAME>.<topic>`) and the listener queue (`<APP_NAME>-in`) |
| `MONGO_URI`              | MongoDB connection URI                                                      |
| `MONGO_DATABASE`         | MongoDB database name                                                       |
| `MONGO_POSTS_COLLECTION` | MongoDB posts collection name                                               |
| `POSTGRES_DSN`           | PostgreSQL DSN                                                              |
| `RABBITMQ_URI`           | RabbitMQ connection URI                                                     |

<br>

---

Made by [Alessandro Massarotti Jr](https://github.com/Alessandro-Massarotti-Jr) 🤖
