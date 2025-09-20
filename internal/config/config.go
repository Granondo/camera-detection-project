package config

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
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
	log.Println("📋 === ЗАГРУЗКА КОНФИГУРАЦИИ ===")
	
	// Пытаемся загрузить .env файл
	log.Println("📄 Попытка загрузки .env файла...")
	if err := LoadEnvFile(".env"); err != nil {
		log.Printf("⚠️ Ошибка загрузки .env файла: %v", err)
	} else {
		log.Println("✅ .env файл обработан успешно")
	}

	log.Println("🔧 Загружаем переменные окружения...")
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
		return nil, fmt.Errorf("❌ RTSP_URL не может быть пустым")
	}

	// Создаем директорию для вывода если она не существует
	if cfg.Camera.SaveFrames {
		log.Printf("📁 Создаем директорию: %s", cfg.Camera.OutputDir)
		if err := os.MkdirAll(cfg.Camera.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("❌ не удалось создать директорию %s: %v", cfg.Camera.OutputDir, err)
		}
		log.Println("✅ Директория для вывода готова")
	}

	log.Printf("✅ === КОНФИГУРАЦИЯ ЗАГРУЖЕНА ===")
	log.Printf("🎯 RTSP URL: %s", maskPassword(cfg.Camera.RTSPUrl))
	log.Printf("👤 Username: %s", cfg.Camera.Username)
	log.Printf("🔒 Password: %s", maskString(cfg.Camera.Password))
	log.Printf("⏱️ Timeout: %d сек", cfg.Camera.Timeout)
	log.Printf("🎞️ Frame Rate: каждый %d кадр", cfg.Camera.FrameRate)
	log.Printf("💾 Save Frames: %v", cfg.Camera.SaveFrames)
	log.Printf("📂 Output Dir: %s", cfg.Camera.OutputDir)

	return cfg, nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		log.Printf("🔧 ENV %s = '%s' (из окружения)", key, value)
		return value
	}
	log.Printf("⚙️ ENV %s = '%s' (по умолчанию)", key, defaultValue)
	return defaultValue
}

// getEnvInt возвращает значение переменной окружения как int или значение по умолчанию
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value != "" {
		if intVal, err := parseInt(value); err == nil {
			log.Printf("🔧 ENV %s = %d (из окружения: '%s')", key, intVal, value)
			return intVal
		}
		log.Printf("⚠️ ENV %s = '%s' (неверный формат, используем по умолчанию: %d)", key, value, defaultValue)
	} else {
		log.Printf("⚙️ ENV %s = %d (по умолчанию)", key, defaultValue)
	}
	return defaultValue
}

// getEnvBool возвращает значение переменной окружения как bool или значение по умолчанию
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value != "" {
		result := value == "true" || value == "1"
		log.Printf("🔧 ENV %s = %v (из окружения: '%s')", key, result, value)
		return result
	}
	log.Printf("⚙️ ENV %s = %v (по умолчанию)", key, defaultValue)
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

// maskString маскирует строку для безопасного логирования
func maskString(s string) string {
	if s == "" {
		return "(пустой)"
	}
	if len(s) <= 2 {
		return "***"
	}
	return s[:1] + "***" + s[len(s)-1:]
}

// maskPassword маскирует пароль в URL
func maskPassword(url string) string {
	if !strings.Contains(url, "@") {
		return url
	}
	// Простая замена пароля в URL
	re := regexp.MustCompile(`://([^:]+):([^@]+)@`)
	return re.ReplaceAllString(url, "://$1:***@")
}