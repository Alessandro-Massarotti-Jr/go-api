# Build
from golang:1.26.4
WORKDIR /build
COPY ./ ./
RUN go mod download
RUN go build -o main cmd/http/main.go
RUN go build -o listener cmd/listener/main.go


# final image
from scratch
COPY --from=0 /build/main /main
COPY --from=0 /build/listener /listener
# Run the RabbitMQ listener with the same image by overriding the command: ["/listener"]
CMD ["/main"]