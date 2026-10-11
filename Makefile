test:
	go test ./internal/entities/... ./internal/usecases/... -cover

run-http:
	go run cmd/http/main.go