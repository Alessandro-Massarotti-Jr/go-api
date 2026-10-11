# Build
FROM golang:1.26.4
WORKDIR /build
COPY go.mod go.sum ./
ENV CGO_ENABLED=0
RUN go mod download
COPY . .
RUN go build -o main cmd/http/main.go
RUN go build -o listener cmd/listener/main.go


# final image
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=0 /build/main /main
COPY --from=0 /build/listener /listener
# Run the RabbitMQ listener with the same image by overriding the command: ["/listener"]
CMD ["/main"]