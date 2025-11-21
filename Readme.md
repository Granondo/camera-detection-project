# Camera Detection Project with FFmpeg

Система видеонаблюдения с детекцией объектов для IP-камер Tapo с поддержкой RTSP, использующая FFmpeg вместо OpenCV для лучшей совместимости с macOS.

## ✨ Особенности

- 📹 Подключение к камерам Tapo через RTSP
- 🎯 Детекция объектов с использованием YOLOv8
- 💾 Сохранение видеозаписей и кадров
- 🗄️ PostgreSQL для хранения метаданных
- 🚀 Redis для кэширования
- 📊 ClickHouse для аналитики временных рядов
- 🐳 Docker поддержка
- ⚙️ Гибкая конфигурация через переменные окружения
- 🍎 Полная совместимость с macOS
- ⚡ Высокая производительность благодаря FFmpeg

## 📋 Требования

- Go 1.21+
- FFmpeg 4.x+ (вместо OpenCV)
- Камера Tapo с поддержкой RTSP

## 🚀 Установка

### macOS (рекомендуется)
```bash
# Установка FFmpeg через Homebrew
make install-ffmpeg-mac

# Клонирование проекта
git clone https://github.com/Granondo/camera-detection-project
cd camera-detection-project
make install
```

### Ubuntu/Debian
```bash
# Полная установка
make install

# Или пошагово:
make setup
make install-deps
```

## ⚙️ Конфигурация

### Настройка камеры Tapo C220
1. Откройте приложение Tapo
2. Перейдите в настройки камеры  
3. Включите RTSP Stream
4. Запомните данные для подключения

### Переменные окружения
```bash
# Создание файла конфигурации
make env-example
cp .env.example .env

# Отредактируйте .env с вашими данными:
# RTSP_URL=rtsp://192.168.1.100:554/stream1
# CAMERA_USERNAME=admin
# CAMERA_PASSWORD=your_password
```

## 🏃‍♂️ Запуск

```bash
# Тест подключения
make test-connection

# Захват тестового кадра
make capture-frame

# Обычный запуск
make run

# Запуск в режиме разработки
make run-dev

# Простой режим (без детекции)
make run-simple
```

### Docker
```bash
# Сборка и запуск
make docker-compose-build

# Или обычный Docker
make docker-build
make docker-run
```

## 📊 Переменные конфигурации

| Переменная | Описание | По умолчанию |
|---|---|---|
| `RTSP_URL` | RTSP URL камеры | `rtsp://192.168.1.100:554/stream1` |
| `CAMERA_USERNAME` | Имя пользователя | `admin` |
| `CAMERA_PASSWORD` | Пароль | (пустой) |
| `CAMERA_TIMEOUT` | Таймаут подключения (сек) | `30` |
| `FRAME_RATE` | Извлекать кадр каждые N секунд | `5` |
| `SAVE_FRAMES` | Сохранять кадры на диск | `true` |
| `OUTPUT_DIR` | Директория для сохранения | `./output` |
| `FFMPEG_PATH` | Путь к FFmpeg | `ffmpeg` |
| `DETECTION_ENABLED` | Включить детекцию | `true` |

## 📊 ClickHouse Аналитика

Система использует ClickHouse для высокопроизводительной аналитики:

### Что хранится в ClickHouse:
- **Детекции объектов** - все обнаруженные объекты с координатами и уверенностью
- **События системы** - логи важных событий (ошибки, детекции, старт/стоп)
- **Метрики производительности** - CPU, память, FPS, latency

### Материализованные представления:
- `detections_hourly` - агрегация детекций по часам
- `detections_daily` - агрегация детекций по дням
- `events_hourly` - агрегация событий по часам

### Retention:
- Детекции - 90 дней
- События - 180 дней
- Метрики - 30 дней

### Запросы к ClickHouse:
```bash
# Подключиться к ClickHouse
make clickhouse-status

# Посмотреть статистику
make clickhouse-stats

# Очистить данные
make clickhouse-clear

# Открыть веб-интерфейс
make clickhouse-web

## 🐰 RabbitMQ Message Queue

Система использует RabbitMQ для асинхронной обработки кадров:

### Преимущества:
- ✅ Асинхронная обработка - камера не ждет детекцию
- ✅ Масштабирование - можно запустить несколько workers
- ✅ Retry механизм - автоматические повторы при ошибках
- ✅ Dead Letter Queue - failed задачи не теряются
- ✅ Load balancing - равномерное распределение нагрузки

### Очереди:
- `frames.to.detect` - кадры на детекцию
- `frames.to.detect.dlq` - failed кадры (Dead Letter Queue)
- `events.notifications` - события для уведомлений

### Команды:
```bash
# Открыть Web UI
make rabbitmq-ui
# http://localhost:15672
# Username: admin
# Password: rabbitmq_password

# Посмотреть статус очередей
make rabbitmq-queues

# Очистить очереди
make rabbitmq-clear

# Масштабировать workers (запустить 5 штук)
make worker-scale N=5

# Логи workers
make worker-logs
```

### Отключение RabbitMQ:

Если хотите использовать синхронную обработку (как раньше):
```bash
# В .env установите:
RABBITMQ_ENABLED=false
```

```

## 🏗️ Архитектура проекта

```
camera-detection-project/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа
├── internal/
│   ├── camera/
│   │   ├── ffmpeg_client.go     # Основной FFmpeg клиент
│   │   ├── simple_camera.go     # Простой клиент (замена OpenCV)
│   │   └── stream_processor.go  # Обработка потока
│   └── config/
│       └── config.go            # Конфигурация
├── output/                      # Сохранённые файлы
│   ├── *.mp4                    # Видеозаписи
│   └── *.jpg                    # Кадры
├── go.mod
├── Makefile
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## 🔧 Компоненты

### FFmpegClient
- Основной клиент для непрерывной записи
- Автоматическое сегментирование видео
- Извлечение кадров для детекции
- Мониторинг процесса FFmpeg

### SimpleCamera  
- Замена OpenCV-based реализации
- Захват отдельных кадров
- Тестирование подключения
- Запись видеосегментов

### StreamProcessor
- Утилиты для обработки потока
- Извлечение кадров
- Тестирование соединения
- Запись сегментов

## 🔧 Преимущества FFmpeg над OpenCV

- **Лучшая совместимость с macOS**: Нет проблем с зависимостями
- **Производительность**: Нативная обработка RTSP потоков
- **Стабильность**: Промышленный стандарт для видео
- **Меньше зависимостей**: Только Go + FFmpeg
- **Гибкость**: Множество кодеков и форматов

## 🛠️ Команды разработки

```bash
make help           # Показать все команды
make check          # Проверка кода (fmt + vet + test)  
make clean          # Очистка
make test           # Тесты
make logs           # Просмотр логов
```

## 🐛 Устранение неисправностей

### Проверка подключения
```bash
# Установите переменную окружения
export RTSP_URL="rtsp://admin:password@192.168.1.100:554/stream1"

# Тест подключения
make test-connection

# Захват тестового кадра
make capture-frame
```

### Частые проблемы

1. **FFmpeg не найден**:
   ```bash
   # macOS
   make install-ffmpeg-mac
   
   # Ubuntu/Debian  
   make install-deps
   ```

2. **Проблемы с RTSP**:
   - Убедитесь, что RTSP включен в настройках камеры
   - Проверьте правильность логина/пароля
   - Попробуйте другой порт (554, 8554)

3. **Права доступа к директории**:
   ```bash
   chmod 755 output/
   ```

4. **Проблемы с Docker**:
   ```bash
   # Проверка версии
   docker --version
   
   # Пересборка образа
   make docker-compose-build
   ```

## 🐳 Docker

```bash
# Быстрый старт с docker-compose
make docker-compose-up

# Или обычный Docker
make docker-build

# Запуск с переменными окружения
RTSP_URL="rtsp://admin:pass@192.168.1.100:554/stream1" \
CAMERA_USERNAME="admin" \
CAMERA_PASSWORD="password" \
make docker-run
```

## 🚧 Планы развития

- [ ] Интеграция YOLO для детекции объектов
- [ ] Web интерфейс для мониторинга
- [ ] Система уведомлений (email, Telegram)
- [ ] База данных для метаданных  
- [ ] Поддержка множественных камер
- [ ] Детекция движения
- [ ] Облачное хранение (S3, Google Cloud)
- [ ] REST API для управления
- [ ] Мобильное приложение

## 🎯 Интеграция детекции объектов

Проект готов для интеграции различных моделей детекции:

```go
// В internal/camera/ffmpeg_client.go
func (c *FFmpegClient) detectObjects() {
    // Здесь можно добавить:
    // - YOLO detection
    // - TensorFlow models
    // - OpenCV DNN
    // - Cloud AI APIs (Google Vision, AWS Rekognition)
    
    log.Println("Running object detection...")
}
```

Примеры интеграции:
- **YOLO**: Через Go bindings или REST API
- **TensorFlow**: Через TensorFlow Go API
- **Cloud APIs**: HTTP запросы к сервисам AI

## 📝 Лицензия

MIT License

## 🆘 Поддержка

При возникновении проблем:
1. Проверьте раздел "Устранение неисправностей"
2. Убедитесь, что FFmpeg установлен: `ffmpeg -version`
3. Протестируйте подключение: `make test-connection`
4. Проверьте логи: `make logs`
5. Создайте issue с описанием проблемы

## 🤝 Вклад в проект

1. Fork проекта
2. Создайте feature branch: `git checkout -b feature/amazing-feature`
3. Commit изменения: `git commit -m 'Add amazing feature'`
4. Push в branch: `git push origin feature/amazing-feature`
5. Создайте Pull Request
