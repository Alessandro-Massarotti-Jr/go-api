test:
	go test ./internal/entities/... ./internal/usecases/... -cover

run-http:
	go run cmd/http/main.go

run-listener:
	go run cmd/listener/main.go

swagger:
	swag init -g cmd/http/main.go -o docs --parseInternal
	swag fmt
