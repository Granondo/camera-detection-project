# Используем официальный Go образ с OpenCV
FROM gocv/opencv:4.8.0-ubuntu

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Создаем директорию для вывода
RUN mkdir -p /app/output

# Собираем приложение
RUN go build -o main ./cmd/server

# Создаем пользователя для безопасности
RUN useradd -u 1001 appuser

# Меняем владельца файлов
RUN chown -R appuser:appuser /app

# Переключаемся на созданного пользователя
USER appuser

# Открываем порт (если понадобится веб-интерфейс)
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]