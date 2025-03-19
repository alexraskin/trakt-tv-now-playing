# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for private repo access
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create data directory, user, and initial token file with proper permissions
RUN mkdir -p /app/data && \
    chmod 777 /app/data && \
    adduser -D -u 1000 appuser && \
    chown appuser:appuser /app/data && \
    echo '{}' > /app/data/token.json && \
    chown appuser:appuser /app/data/token.json && \
    chmod 666 /app/data/token.json

# Copy the binary from builder
COPY --from=builder /app/main .

# Switch to appuser
USER appuser

# Set token file path to use the persistent volume
ENV TOKEN_FILE=/app/data/token.json

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./main"] 