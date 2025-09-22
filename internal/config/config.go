package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	RTSPURL         string
	Username        string
	Password        string
	CameraTimeout   time.Duration
	FrameRate       int
	SaveFrames      bool
	OutputDir       string
	FFmpegPath      string
	DetectionEnabled bool
}

func Load() (*Config, error) {
	// Try to load .env file if it exists
	loadEnvFile()

	cfg := &Config{
		RTSPURL:         getEnv("RTSP_URL", "rtsp://192.168.1.100:554/stream1"),
		Username:        getEnv("CAMERA_USERNAME", "admin"),
		Password:        getEnv("CAMERA_PASSWORD", ""),
		CameraTimeout:   getDurationEnv("CAMERA_TIMEOUT", 30*time.Second),
		FrameRate:       getIntEnv("FRAME_RATE", 5),
		SaveFrames:      getBoolEnv("SAVE_FRAMES", true),
		OutputDir:       getEnv("OUTPUT_DIR", "./output"),
		FFmpegPath:      getEnv("FFMPEG_PATH", "ffmpeg"),
		DetectionEnabled: getBoolEnv("DETECTION_ENABLED", true),
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