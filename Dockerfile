# Build stage
FROM golang:1.21-alpine AS builder

# Install git for downloading dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and su-exec for user switching
RUN apk --no-cache add ca-certificates tzdata su-exec

# Create a non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy entrypoint script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Create data directory and set initial permissions
RUN mkdir -p data && chown -R appuser:appuser /app

# Set entrypoint to handle permissions
ENTRYPOINT ["/docker-entrypoint.sh"]

# Expose port (not needed for Telegram bot but good practice)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD pgrep main || exit 1

# Run the binary
CMD ["./main"] 