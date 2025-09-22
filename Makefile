.PHONY: help setup install-deps build run run-dev test clean docker-build docker-run install-ffmpeg-mac

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Setup project dependencies
	go mod tidy
	mkdir -p output
	mkdir -p logs

install-deps: ## Install system dependencies
	@echo "Installing FFmpeg..."
	@if command -v brew >/dev/null 2>&1; then \
		echo "Installing via Homebrew (macOS)..."; \
		brew install ffmpeg; \
	elif command -v apt-get >/dev/null 2>&1; then \
		echo "Installing via apt (Ubuntu/Debian)..."; \
		sudo apt-get update && sudo apt-get install -y ffmpeg; \
	elif command -v yum >/dev/null 2>&1; then \
		echo "Installing via yum (CentOS/RHEL)..."; \
		sudo yum install -y ffmpeg; \
	else \
		echo "Please install FFmpeg manually from https://ffmpeg.org/"; \
		echo "Or use: make install-ffmpeg-mac for macOS"; \
	fi

install-ffmpeg-mac: ## Install FFmpeg on macOS
	@if command -v brew >/dev/null 2>&1; then \
		brew install ffmpeg; \
	else \
		echo "Homebrew not found. Please install Homebrew first:"; \
		echo "/bin/bash -c \"\$$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""; \
	fi

build: ## Build the application
	@echo "Building camera detection service..."
	go build -o bin/camera-detection cmd/server/main.go

run: build ## Run the application
	@echo "Starting camera detection service..."
	./bin/camera-detection

run-dev: ## Run in development mode
	@echo "Running in development mode..."
	go run cmd/server/main.go

run-simple: ## Run with simple camera mode
	@echo "Running simple camera test..."
	DETECTION_ENABLED=false go run cmd/server/main.go

test: ## Run tests
	go test ./...

check: ## Run code quality checks
	@echo "Running code quality checks..."
	go fmt ./...
	go vet ./...
	go test ./...

clean: ## Clean build artifacts and output
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf output/*.mp4
	rm -rf output/*.jpg
	rm -rf logs/*
	go clean

test-connection: ## Test RTSP connection
	@echo "Testing RTSP connection..."
	@if [ -z "$$RTSP_URL" ]; then \
		echo "Please set RTSP_URL environment variable"; \
		echo "Example: export RTSP_URL='rtsp://admin:password@192.168.1.100:554/stream1'"; \
		exit 1; \
	fi
	ffprobe -rtsp_transport tcp -i "$$RTSP_URL" -t 1 -f null -

capture-frame: ## Capture a single frame for testing
	@echo "Capturing test frame..."
	@if [ -z "$$RTSP_URL" ]; then \
		echo "Please set RTSP_URL environment variable"; \
		exit 1; \
	fi
	mkdir -p output
	ffmpeg -rtsp_transport tcp -i "$$RTSP_URL" -vframes 1 -q:v 2 -y output/test_frame.jpg

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t camera-detection:latest .

docker-run: ## Run in Docker container
	@echo "Running in Docker..."
	docker run --rm \
		-e RTSP_URL="$$RTSP_URL" \
		-e CAMERA_USERNAME="$$CAMERA_USERNAME" \
		-e CAMERA_PASSWORD="$$CAMERA_PASSWORD" \
		-e SAVE_FRAMES="$$SAVE_FRAMES" \
		-e DETECTION_ENABLED="$$DETECTION_ENABLED" \
		-v $$(pwd)/output:/app/output \
		camera-detection:latest

docker-compose-up: ## Start with docker-compose
	docker-compose up

docker-compose-build: ## Build and start with docker-compose
	docker-compose up --build

logs: ## Show recent logs
	@if [ -d "logs" ]; then \
		tail -f logs/*.log; \
	else \
		echo "No log directory found"; \
	fi

env-example: ## Create .env example file
	@echo "Creating .env.example file..."
	@echo "# RTSP Configuration" > .env.example
	@echo "RTSP_URL=rtsp://192.168.1.100:554/stream1" >> .env.example
	@echo "CAMERA_USERNAME=admin" >> .env.example
	@echo "CAMERA_PASSWORD=your_password" >> .env.example
	@echo "" >> .env.example
	@echo "# Processing Configuration" >> .env.example
	@echo "CAMERA_TIMEOUT=30" >> .env.example
	@echo "FRAME_RATE=5" >> .env.example
	@echo "SAVE_FRAMES=true" >> .env.example
	@echo "OUTPUT_DIR=./output" >> .env.example
	@echo "" >> .env.example
	@echo "# FFmpeg Configuration" >> .env.example
	@echo "FFMPEG_PATH=ffmpeg" >> .env.example
	@echo "DETECTION_ENABLED=true" >> .env.example
	@echo ".env.example created successfully!"

install: setup install-deps ## Complete installation
	@echo "Installation completed!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Copy .env.example to .env and configure your camera settings"
	@echo "2. Run 'make test-connection' to test your RTSP connection"
	@echo "3. Run 'make run' to start the application"