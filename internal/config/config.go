package config

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type DetectionServiceConfig struct {
	URL                 string
	Timeout             time.Duration
	MaxRetries          int
	ConfidenceThreshold float64
}

type Config struct {
	// Camera settings
	RTSPURL          string
	Username         string
	Password         string
	CameraTimeout    time.Duration
	FrameRate        int
	SaveFrames       bool
	OutputDir        string
	FFmpegPath       string
	DetectionEnabled bool

	// Database settings
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseSSLMode  string

	// Redis settings
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int
	RedisCacheTTL time.Duration

	// ClickHouse settings
	ClickHouseHost     string
	ClickHousePort     int
	ClickHouseDatabase string
	ClickHouseUsername string
	ClickHousePassword string
	ClickHouseDebug    bool

	// RabbitMQ settings
	RabbitMQHost     string
	RabbitMQPort     int
	RabbitMQUser     string
	RabbitMQPassword string
	RabbitMQVHost    string
	RabbitMQEnabled  bool

	// Elasticsearch settings
	ElasticsearchEnabled bool
	ElasticsearchURL     string
	ElasticsearchIndex   string

	// Detection settings
	DetectionService DetectionServiceConfig
}

func Load() (*Config, error) {
	// Try to load .env file if it exists
	loadEnvFile()

	cfg := &Config{
		// Camera configuration
		RTSPURL:          getEnv("RTSP_URL", "rtsp://192.168.1.100:554/stream1"),
		Username:         getEnv("CAMERA_USERNAME", "admin"),
		Password:         getEnv("CAMERA_PASSWORD", ""),
		CameraTimeout:    getDurationEnv("CAMERA_TIMEOUT", 30*time.Second),
		FrameRate:        getIntEnv("FRAME_RATE", 5),
		SaveFrames:       getBoolEnv("SAVE_FRAMES", true),
		OutputDir:        getEnv("OUTPUT_DIR", "./output"),
		FFmpegPath:       getEnv("FFMPEG_PATH", "ffmpeg"),
		DetectionEnabled: getBoolEnv("DETECTION_ENABLED", true),

		// Database configuration
		DatabaseHost:     getEnv("DATABASE_HOST", "localhost"),
		DatabasePort:     getIntEnv("DATABASE_PORT", 5432),
		DatabaseUser:     getEnv("DATABASE_USER", "postgres"),
		DatabasePassword: getEnv("DATABASE_PASSWORD", "postgres"),
		DatabaseName:     getEnv("DATABASE_NAME", "surveillance"),
		DatabaseSSLMode:  getEnv("DATABASE_SSL_MODE", "disable"),

		// Redis configuration
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getIntEnv("REDIS_PORT", 6379),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getIntEnv("REDIS_DB", 0),
		RedisCacheTTL: getDurationEnv("REDIS_CACHE_TTL", 300*time.Second),

		// ClickHouse configuration
		ClickHouseHost:     getEnv("CLICKHOUSE_HOST", "localhost"),
		ClickHousePort:     getIntEnv("CLICKHOUSE_PORT", 9000),
		ClickHouseDatabase: getEnv("CLICKHOUSE_DATABASE", "surveillance"),
		ClickHouseUsername: getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),
		ClickHouseDebug:    getBoolEnv("CLICKHOUSE_DEBUG", false),

		// RabbitMQ configuration
		RabbitMQHost:     getEnv("RABBITMQ_HOST", "localhost"),
		RabbitMQPort:     getIntEnv("RABBITMQ_PORT", 5672),
		RabbitMQUser:     getEnv("RABBITMQ_USER", "admin"),
		RabbitMQPassword: getEnv("RABBITMQ_PASSWORD", "rabbitmq_password"),
		RabbitMQVHost:    getEnv("RABBITMQ_VHOST", "/"),
		RabbitMQEnabled:  getBoolEnv("RABBITMQ_ENABLED", true),

		// Elasticsearch configuration
		ElasticsearchEnabled: getBoolEnv("ELASTICSEARCH_ENABLED", true),
		ElasticsearchURL:     getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		ElasticsearchIndex:   getEnv("ELASTICSEARCH_INDEX", "surveillance-events"),

		DetectionService: DetectionServiceConfig{
			URL:                 getEnv("DETECTION_SERVICE_URL", "http://detection-service:5000"),
			Timeout:             getDurationEnv("DETECTION_SERVICE_TIMEOUT", 30*time.Second),
			MaxRetries:          getIntEnv("DETECTION_SERVICE_MAX_RETRIES", 3),
			ConfidenceThreshold: getFloatEnv("DETECTION_CONFIDENCE_THRESHOLD", 0.5),
		},
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadEnvFile loads .env file if it exists
func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		return // .env file doesn't exist, skip
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Set environment variable only if not already set
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
				displayValue := value
				if strings.Contains(strings.ToLower(key), "password") {
					displayValue = maskValue(value)
				}
				log.Printf("📝 Загружен %s = %s", key, displayValue)
			}
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value + "s"); err == nil {
			return duration
		}
	}
	return defaultValue
}

func maskValue(value string) string {
	if value == "" {
		return "(пустое)"
	}
	return value[:1] + "***" + value[len(value)-1:]
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
