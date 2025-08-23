# Multi-stage build for Go backend
FROM golang:1.24.3-alpine AS builder

# Install necessary packages for building
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o main cmd/api/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o migration cmd/migration/main.go

# Production stage
FROM alpine:latest

# Install necessary runtime packages
RUN apk add --no-cache ca-certificates curl

# Create app user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/migration .

# Create data directory and set ownership
RUN mkdir -p /data && chown -R appuser:appgroup /app /data

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8288

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8288/api/health || exit 1

# Run the binary
CMD ["./main"]