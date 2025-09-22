FROM golang:1.21-alpine AS builder

# Install FFmpeg and build dependencies
RUN apk add --no-cache \
    ffmpeg \
    gcc \
    musl-dev \
    git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o camera-detection cmd/server/main.go

# Final stage
FROM alpine:latest

# Install FFmpeg runtime
RUN apk add --no-cache ffmpeg

# Create app user
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/camera-detection .

# Create directories and set permissions
RUN mkdir -p output logs && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pgrep camera-detection || exit 1

# Default environment variables
ENV RTSP_URL=rtsp://192.168.1.100:554/stream1
ENV CAMERA_USERNAME=admin
ENV CAMERA_PASSWORD=""
ENV SAVE_FRAMES=true
ENV OUTPUT_DIR=/app/output
ENV DETECTION_ENABLED=true

EXPOSE 8080

CMD ["./camera-detection"]