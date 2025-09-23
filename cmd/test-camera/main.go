package main

import (
	"flag"
	"log"
	"os"

	"camera-detection-project/internal/camera"
	"camera-detection-project/internal/config"
)

func main() {
	var (
		testConnection = flag.Bool("connection", false, "Test RTSP connection only")
		captureFrame   = flag.Bool("frame", false, "Capture a single frame")
		recordSegment  = flag.Int("record", 0, "Record video segment (seconds)")
		rtspURL        = flag.String("url", "", "RTSP URL (overrides config)")
	)
	flag.Parse()

	log.Println("🔧 Camera Test Utility")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override RTSP URL if provided
	if *rtspURL != "" {
		cfg.RTSPURL = *rtspURL
		cfg.Username = "" // Clear credentials when URL is overridden
		cfg.Password = ""
	}

	// Determine what to test
	switch {
	case *testConnection:
		testConnectionOnly(cfg)
	case *captureFrame:
		captureFrameOnly(cfg)
	case *recordSegment > 0:
		recordVideoSegment(cfg, *recordSegment)
	default:
		runFullTest(cfg)
	}
}

func testConnectionOnly(cfg *config.Config) {
	log.Println("🔌 Testing RTSP connection...")
	
	if err := camera.TestRTSPConnectionWithConfig(cfg); err != nil {
		log.Fatalf("❌ Connection test failed: %v", err)
	}
	
	log.Println("✅ Connection test successful!")
}

func captureFrameOnly(cfg *config.Config) {
	log.Println("📸 Capturing single frame...")
	
	outputPath := "output/test_frame.jpg"
	if err := camera.ExtractFrameWithConfig(cfg, outputPath); err != nil {
		log.Fatalf("❌ Frame capture failed: %v", err)
	}
	
	log.Printf("✅ Frame captured: %s", outputPath)
}

func recordVideoSegment(cfg *config.Config, duration int) {
	log.Printf("🎥 Recording %d second video segment...", duration)
	
	outputPath := "output/test_recording.mp4"
	rtspURL := buildRTSPURL(cfg.RTSPURL, cfg.Username, cfg.Password)
	
	if err := camera.RecordVideoSegment(rtspURL, outputPath, duration); err != nil {
		log.Fatalf("❌ Video recording failed: %v", err)
	}
	
	log.Printf("✅ Video recorded: %s", outputPath)
}

func runFullTest(cfg *config.Config) {
	log.Println("🚀 Running full camera test...")
	
	if err := camera.QuickCameraTest(cfg); err != nil {
		log.Fatalf("❌ Camera test failed: %v", err)
	}
	
	log.Println("🎉 All tests completed successfully!")
}

// buildRTSPURL constructs RTSP URL with credentials
func buildRTSPURL(baseURL, username, password string) string {
	if username == "" || password == "" {
		return baseURL
	}
	
	if baseURL[:7] == "rtsp://" {
		return "rtsp://" + username + ":" + password + "@" + baseURL[7:]
	}
	return baseURL
}

func init() {
	// Ensure output directory exists
	if err := os.MkdirAll("output", 0755); err != nil {
		log.Printf("Warning: could not create output directory: %v", err)
	}
}