package config

import (
	"os"
	"strconv"
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