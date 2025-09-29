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
	go build -o camera-detection cmd/server/main.go

run: build ## Run the application
	@echo "Starting camera detection service..."
	./camera-detection

run-dev: ## Run in development mode
	@echo "Running in development mode..."
	go run cmd/server/main.go

test: ## Run tests
	go test ./...

check: ## Run code quality checks
	@echo "Running code quality checks..."
	go fmt ./...
	go vet ./...
	go test ./...

clean: ## Clean build artifacts and output
	@echo "Cleaning up..."
	rm -f camera-detection
	rm -rf output/*.mp4
	rm -rf output/*.jpg
	rm -rf logs/*
	go clean

test-connection: ## Test RTSP connection
	@echo "Testing RTSP connection..."
	@if [ -z "$RTSP_URL" ]; then \
		echo "Please set RTSP_URL environment variable"; \
		echo "Example: export RTSP_URL='rtsp://admin:password@192.168.1.100:554/stream1'"; \
		exit 1; \
	fi
	ffprobe -rtsp_transport tcp -i "$RTSP_URL" -t 1 -f null -

quick-test: ## Quick camera functionality test
	@echo "Running quick camera test..."
	go run -ldflags="-X main.mode=test" cmd/server/main.go || \
	go run cmd/test-camera/main.go 2>/dev/null || \
	echo "Quick test requires .env file with camera configuration"

capture-frame: ## Capture a single frame for testing
	@echo "Capturing test frame..."
	@if [ -z "$RTSP_URL" ]; then \
		echo "Please set RTSP_URL environment variable"; \
		exit 1; \
	fi
	mkdir -p output
	ffmpeg -rtsp_transport tcp -i "$RTSP_URL" -vframes 1 -q:v 2 -y output/test_frame.jpg

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

install: setup install-deps ## Complete installation
	@echo "Installation completed!"
	@echo ""
	@echo "Next steps:"
	@if [ ! -f .env ]; then \
		echo "1. Create .env file with your camera and database settings:"; \
		echo "   cp .env.example .env  # or create manually"; \
		echo "   Edit .env with your camera IP, credentials, and database settings"; \
	else \
		echo "1. ✅ .env file already exists"; \
	fi
	@echo "2. Run 'make db-start' to start PostgreSQL"
	@echo "3. Run 'make db-migrate' to create database tables"
	@echo "4. Run 'make test-connection' to test your RTSP connection"
	@echo "5. Run 'make run' to start the application"

check-env: ## Check if .env file exists
	@if [ ! -f .env ]; then \
		echo "❌ .env file not found!"; \
		echo ""; \
		echo "Please create .env file with the following variables:"; \
		echo ""; \
		echo "# Camera Configuration"; \
		echo "RTSP_URL=rtsp://192.168.1.71:554/stream1"; \
		echo "CAMERA_USERNAME=your_email@gmail.com"; \
		echo "CAMERA_PASSWORD=your_password"; \
		echo ""; \
		echo "# Processing Configuration"; \
		echo "CAMERA_TIMEOUT=30"; \
		echo "FRAME_RATE=5"; \
		echo "SAVE_FRAMES=true"; \
		echo "OUTPUT_DIR=./output"; \
		echo ""; \
		echo "# FFmpeg Configuration"; \
		echo "FFMPEG_PATH=ffmpeg"; \
		echo "DETECTION_ENABLED=true"; \
		echo ""; \
		echo "# Database Configuration"; \
		echo "DATABASE_HOST=localhost"; \
		echo "DATABASE_PORT=5432"; \
		echo "DATABASE_USER=postgres"; \
		echo "DATABASE_PASSWORD=postgres"; \
		echo "DATABASE_NAME=surveillance"; \
		echo "DATABASE_SSL_MODE=disable"; \
		echo ""; \
		exit 1; \
	else \
		echo "✅ .env file exists"; \
	fi

# Database commands
db-start: ## Start PostgreSQL with Docker Compose
	@echo "Starting PostgreSQL database..."
	docker-compose up -d postgres

db-stop: ## Stop PostgreSQL
	@echo "Stopping PostgreSQL database..."
	docker-compose stop postgres

db-migrate: ## Run database migrations (create tables)
	@echo "Creating database tables..."
	go run cmd/migrate/main.go

db-reset: ## Reset database (WARNING: destroys all data)
	@echo "WARNING: This will destroy all data!"
	@echo "Press Ctrl+C to cancel, or wait 5 seconds to continue..."
	@sleep 5
	docker-compose down postgres
	docker volume rm $(docker volume ls -q | grep postgres) 2>/dev/null || true
	docker-compose up -d postgres
	@sleep 5
	make db-migrate

db-stats: ## Show database statistics
	@echo "Database statistics:"
	go run cmd/db-stats/main.go

db-shell: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U postgres -d surveillance

pgadmin: ## Start pgAdmin web interface
	@echo "Starting pgAdmin..."
	docker-compose up -d pgadmin
	@echo "pgAdmin available at: http://localhost:8080"

# Detection service commands
detection-build: ## Build detection service
	@echo "Building detection service..."
	docker build -t yolo-detection-service ./detection_service

detection-run: ## Run detection service standalone
	@echo "Starting detection service..."
	docker run --rm \
		-p 5001:5000 \
		-v $(pwd)/output:/app/data \
		-e CONFIDENCE_THRESHOLD=0.5 \
		yolo-detection-service

detection-test: ## Test detection service
	@echo "Testing detection service..."
	@if [ -z "$(IMAGE_PATH)" ]; then \
		echo "Please provide IMAGE_PATH: make detection-test IMAGE_PATH=output/frame_123.jpg"; \
		exit 1; \
	fi
	curl -X POST http://localhost:5001/detect \
		-H "Content-Type: application/json" \
		-d '{"image_path": "$(IMAGE_PATH)"}' \
		| jq .

detection-health: ## Check detection service health
	@echo "Checking detection service health..."
	curl -s http://localhost:5001/health | jq .

# Full system commands with detection
start-full: ## Start full system with detection
	@echo "Starting full surveillance system..."
	docker-compose up --build

stop-full: ## Stop full system
	@echo "Stopping full surveillance system..."
	docker-compose down

restart-detection: ## Restart only detection service
	@echo "Restarting detection service..."
	docker-compose restart detection-service

logs-detection: ## Show detection service logs
	@echo "Detection service logs:"
	docker-compose logs -f detection-service

db-clear: ## Clear all data from database (keep tables)
	@echo "⚠️  Clearing all data from database..."
	docker-compose exec postgres psql -U postgres -d surveillance -c "\
		TRUNCATE TABLE detections CASCADE; \
		TRUNCATE TABLE frames CASCADE; \
		TRUNCATE TABLE recordings CASCADE; \
		TRUNCATE TABLE events CASCADE; \
		TRUNCATE TABLE system_stats CASCADE; \
		TRUNCATE TABLE cameras CASCADE; \
		ALTER SEQUENCE cameras_id_seq RESTART WITH 1; \
		ALTER SEQUENCE recordings_id_seq RESTART WITH 1; \
		ALTER SEQUENCE frames_id_seq RESTART WITH 1; \
		ALTER SEQUENCE detections_id_seq RESTART WITH 1; \
		ALTER SEQUENCE events_id_seq RESTART WITH 1; \
		ALTER SEQUENCE system_stats_id_seq RESTART WITH 1;"
	@echo "✅ Database cleared"

output-clear: ## Clear output directory
	@echo "🗑️  Clearing output directory..."
	rm -rf output/*.jpg output/*.mp4 output/*.h264
	@echo "✅ Output cleared"

clear-all: db-clear output-clear ## Clear both database and output
	@echo "🎉 Everything cleared!"