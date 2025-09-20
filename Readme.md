# Camera Detection Project

Система видеонаблюдения с детекцией объектов для IP-камер Tapo с поддержкой RTSP.

## Возможности

- 📹 Подключение к камерам Tapo через RTSP
- 🎯 Детекция людей и животных (в разработке)
- 💾 Сохранение кадров и метаданных
- 🐳 Docker поддержка
- ⚙️ Гибкая конфигурация через переменные окружения

## Требования

### Основные зависимости
- Go 1.21+
- OpenCV 4.x
- Камера Tapo с поддержкой RTSP

### Установка OpenCV

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install -y libopencv-dev
# или используйте Makefile
make install-opencv
```

#### macOS
```bash
brew install opencv
# или используйте Makefile
make install-opencv-mac
```

## Быстрый старт

### 1. Клонирование и подготовка

```bash
git clone <repository-url>
cd camera-detection-project
make setup
make install-deps
```

### 2. Настройка камеры

Убедитесь, что ваша камера Tapo C220 настроена для RTSP:

1. Откройте приложение Tapo
2. Перейдите в настройки камеры
3. Включите RTSP Stream
4. Запомните данные для подключения

### 3. Конфигурация

Создайте файл `.env` и установите переменные окружения

### 4. Запуск

```bash
# Обычный запуск
make run

# Запуск с настройками разработки
make run-dev

# Запуск в Docker
make docker-run
```

## Конфигурация

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `RTSP_URL` | RTSP URL камеры | `rtsp://192.168.1.100:554/stream1` |
| `CAMERA_USERNAME` | Имя пользователя | `admin` |
| `CAMERA_PASSWORD` | Пароль | _(пустой)_ |
| `CAMERA_TIMEOUT` | Таймаут подключения (сек) | `30` |
| `FRAME_RATE` | Обрабатывать каждый N-й кадр | `5` |
| `SAVE_FRAMES` | Сохранять кадры на диск | `true` |
| `OUTPUT_DIR` | Директория для сохранения | `./output` |

## Структура проекта

```
camera-detection-project/
├── cmd/
│   └── server/             # Точка входа приложения
│       └── main.go
├── internal/
│   ├── camera/            # RTSP клиент и обработка видео
│   │   ├── camera.go      # Продвинутый RTSP клиент  
│   │   └── simple_camera.go # Простой клиент через OpenCV
│   ├── config/            # Конфигурация
│   │   └── config.go
│   ├── detector/          # ML модели детекции (планируется)
│   └── storage/           # База данных (планируется)
├── output/                # Сохраненные кадры
├── go.mod
├── go.sum  
├── Dockerfile
├── Makefile
└── README.md
```

## Поиск проблем

### Проблемы с подключением к камере

1. **Проверьте сетевое подключение:**
   ```bash
   ping 192.168.1.100  # замените на IP вашей камеры
   ```

2. **Проверьте RTSP поток:**
   ```bash
   ffplay rtsp://admin:password@192.168.1.100:554/stream1
   ```

3. **Проверьте настройки камеры в приложении Tapo**

### Ошибки сборки

1. **OpenCV не найден:**
   ```bash
   # Ubuntu/Debian
   sudo apt install libopencv-dev pkg-config

   # macOS
   brew install opencv pkg-config
   ```

2. **Go модули:**
   ```bash
   go mod download
   go mod tidy
   ```

## Docker

### Сборка и запуск

```bash
# Сборка образа
make docker-build

# Запуск контейнера
docker run --rm \
  -e RTSP_URL="rtsp://192.168.1.100:554/stream1" \
  -e CAMERA_USERNAME="admin" \
  -e CAMERA_PASSWORD="password" \
  -v $(pwd)/output:/app/output \
  camera-detection-project:latest
```

## Разработка

### Полезные команды

```bash
make help           # Показать все доступные команды
make check          # Проверка кода (fmt + vet + test)
make clean          # Очистка
make test           # Тесты
```

### Следующие шаги

- [ ] Добавить детекцию объектов (YOLO/TensorFlow)
- [ ] Реализовать базу данных для метаданных
- [ ] Создать веб-интерфейс
- [ ] Добавить систему уведомлений
- [ ] Оптимизировать производительность

## Лицензия

MIT License

## Поддержка

Если у вас возникли проблемы:

1. Проверьте раздел "Поиск проблем"
2. Убедитесь, что все зависимости установлены
3. Проверьте логи приложения
4. Создайте issue с описанием проблемы