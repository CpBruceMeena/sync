# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o server ./cmd/server

# Stage 2: Minimal runtime image
FROM alpine:3.19 AS runtime

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]

# Stage 3: Test runner
FROM golang:1.23-alpine AS test

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source (exclude tests? No — we need them!)
COPY . .

# Run unit tests (E2E tests use testcontainers or TEST_DATABASE_URL)
CMD ["go", "test", "./tests/...", "-v", "-count=1"]
