FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build detection worker
RUN go build -o detection_worker cmd/detection_worker/main.go

# Final stage
FROM alpine:latest

RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

WORKDIR /app

COPY --from=builder /app/detection_worker .

# Create directories
RUN mkdir -p output logs && \
    chown -R appuser:appgroup /app

USER appuser

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pgrep detection_worker || exit 1

CMD ["./detection_worker"]