# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o function ./cmd/cloudfunction

# Runtime stage
FROM alpine:latest

# Install FFmpeg (required for video processing)
RUN apk add --no-cache ffmpeg ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/function .

# Expose port
EXPOSE 8080

# Run the function
CMD ["./function"]