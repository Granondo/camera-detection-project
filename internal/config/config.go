package config

import (
	"fmt"
	"os"
)

// Config содержит всю конфигурацию приложения
type Config struct {
	Camera CameraConfig
}

// CameraConfig содержит настройки камеры
type CameraConfig struct {
	RTSPUrl     string
	Username    string
	Password    string
	Timeout     int // секунды
	FrameRate   int // кадров для обработки (каждый N-й кадр)
	SaveFrames  bool
	OutputDir   string
}

// Load загружает конфигурацию из переменных окружения или использует значения по умолчанию
func Load() (*Config, error) {
	cfg := &Config{
		Camera: CameraConfig{
			RTSPUrl:     getEnv("RTSP_URL", "rtsp://192.168.1.100:554/stream1"),
			Username:    getEnv("CAMERA_USERNAME", "admin"),
			Password:    getEnv("CAMERA_PASSWORD", ""),
			Timeout:     getEnvInt("CAMERA_TIMEOUT", 30),
			FrameRate:   getEnvInt("FRAME_RATE", 5), // обрабатывать каждый 5-й кадр
			SaveFrames:  getEnvBool("SAVE_FRAMES", true),
			OutputDir:   getEnv("OUTPUT_DIR", "./output"),
		},
	}

	// Проверяем обязательные параметры
	if cfg.Camera.RTSPUrl == "" {
		return nil, fmt.Errorf("RTSP_URL не может быть пустым")
	}

	// Создаем директорию для вывода если она не существует
	if cfg.Camera.SaveFrames {
		if err := os.MkdirAll(cfg.Camera.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("не удалось создать директорию %s: %v", cfg.Camera.OutputDir, err)
		}
	}

	return cfg, nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt возвращает значение переменной окружения как int или значение по умолчанию
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := parseInt(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvBool возвращает значение переменной окружения как bool или значение по умолчанию
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

// parseInt простая функция для парсинга строки в int
func parseInt(s string) (int, error) {
	result := 0
	for _, digit := range s {
		if digit < '0' || digit > '9' {
			return 0, fmt.Errorf("неверный формат числа")
		}
		result = result*10 + int(digit-'0')
	}
	return result, nil
}