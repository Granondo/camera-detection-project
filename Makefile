# Makefile для Camera Detection Project

.PHONY: build run clean test docker-build docker-run install-deps

# Переменные
BINARY_NAME=camera-detection-project
DOCKER_IMAGE=camera-detection-project:latest

# Сборка приложения
build:
	@echo "Сборка приложения..."
	go build -o $(BINARY_NAME) ./cmd/server

# Запуск приложения
run: build
	@echo "Запуск приложения..."
	./$(BINARY_NAME)

# Запуск с переменными окружения
run-dev: build
	@echo "Запуск в режиме разработки..."
	@if [ -f .env ]; then \
		echo "✅ Используется .env файл"; \
	else \
		echo "⚠️  .env файл не найден, используются значения по умолчанию"; \
		echo "💡 Запустите 'make setup-env' для создания .env файла"; \
	fi
	./$(BINARY_NAME)

# Создание .env файла из примера
setup-env:
	@if [ ! -f .env ]; then \
		echo "# Camera Detection Project - Configuration" > .env; \
		echo "# Отредактируйте значения под ваши настройки" >> .env; \
		echo "" >> .env; \
		echo "RTSP_URL=rtsp://192.168.1.100:554/stream1" >> .env; \
		echo "CAMERA_USERNAME=admin" >> .env; \
		echo "CAMERA_PASSWORD=your_camera_password" >> .env; \
		echo "FRAME_RATE=5" >> .env; \
		echo "SAVE_FRAMES=true" >> .env; \
		echo "OUTPUT_DIR=./output" >> .env; \
		echo "CAMERA_TIMEOUT=30" >> .env; \
		echo "✅ Создан .env файл с настройками по умолчанию"; \
		echo "📝 Отредактируйте .env файл с вашими реальными настройками"; \
	else \
		echo "⚠️  .env файл уже существует"; \
	fi

# Установка зависимостей (требует установленного OpenCV)
install-deps:
	@echo "Установка Go зависимостей..."
	go mod download
	go mod tidy

# Установка OpenCV (Ubuntu/Debian)
install-opencv:
	@echo "Установка OpenCV..."
	sudo apt update
	sudo apt install -y libopencv-dev

# Установка OpenCV (macOS)
install-opencv-mac:
	@echo "Установка OpenCV на macOS..."
	brew install opencv

# Очистка
clean:
	@echo "Очистка..."
	go clean
	rm -f $(BINARY_NAME)
	rm -rf output/

# Тестирование
test:
	@echo "Запуск тестов..."
	go test -v ./...

# Сборка Docker образа
docker-build:
	@echo "Сборка Docker образа..."
	docker build -t $(DOCKER_IMAGE) .

# Запуск в Docker
docker-run: docker-build
	@echo "Запуск в Docker контейнере..."
	docker run --rm \
		-e RTSP_URL="rtsp://192.168.1.100:554/stream1" \
		-e CAMERA_USERNAME="admin" \
		-e CAMERA_PASSWORD="your_password" \
		-v $(PWD)/output:/app/output \
		$(DOCKER_IMAGE)

# Создание директорий
setup:
	@echo "Создание необходимых директорий..."
	mkdir -p output
	mkdir -p cmd/server
	mkdir -p internal/camera
	mkdir -p internal/config
	mkdir -p internal/detector
	mkdir -p internal/storage
	@if [ ! -f .env ]; then \
		echo "📝 Создаю базовый .env файл..."; \
		$(MAKE) setup-env; \
	else \
		echo "✅ .env файл уже существует"; \
	fi

# Проверка форматирования кода
fmt:
	@echo "Форматирование кода..."
	go fmt ./...

# Проверка кода
vet:
	@echo "Проверка кода..."
	go vet ./...

# Все проверки
check: fmt vet test

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  build              - Сборка приложения"
	@echo "  run                - Запуск приложения"
	@echo "  run-dev            - Запуск с переменными окружения"
	@echo "  install-deps       - Установка Go зависимостей"
	@echo "  install-opencv     - Установка OpenCV (Ubuntu/Debian)"
	@echo "  install-opencv-mac - Установка OpenCV (macOS)"
	@echo "  clean              - Очистка собранных файлов"
	@echo "  test               - Запуск тестов"
	@echo "  docker-build       - Сборка Docker образа"
	@echo "  docker-run         - Запуск в Docker"
	  @echo "  setup              - Создание необходимых директорий"
  @echo "  setup-env          - Создать .env файл из примера"
	@echo "  fmt                - Форматирование кода"
	@echo "  vet                - Проверка кода"
	@echo "  check              - Все проверки (fmt + vet + test)"