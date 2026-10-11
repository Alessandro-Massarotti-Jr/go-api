test:
	go test ./... -cover

run-http:
	go run cmd/http/main.go