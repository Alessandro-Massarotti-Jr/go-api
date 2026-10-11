# Build
from golang:1.26.4
WORKDIR /build
COPY ./ ./
RUN go mod download
RUN go build -o main cmd/http/main.go


# final image
from scratch
COPY --from=0 /build/main /main
CMD ["/main"]