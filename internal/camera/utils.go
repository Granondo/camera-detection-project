package camera

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"camera-detection-project/internal/config"
)

// TestRTSPConnection tests RTSP connection using ffprobe
func TestRTSPConnection(rtspURL string) error {
	log.Printf("Testing RTSP connection to: %s", maskPassword(rtspURL))

	args := []string{
		"-rtsp_transport", "tcp",
		"-i", rtspURL,
		"-t", "3",                // Test for 3 seconds
		"-f", "null",             // No output
		"-v", "quiet",            // Quiet output
		"-",
	}

	cmd := exec.Command("ffprobe", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("RTSP connection test failed: %w", err)
	}

	log.Println("RTSP connection test successful")
	return nil
}

// TestRTSPConnectionWithConfig tests connection using config
func TestRTSPConnectionWithConfig(cfg *config.Config) error {
	rtspURL := buildRTSPURL(cfg.RTSPURL, cfg.Username, cfg.Password)
	return TestRTSPConnection(rtspURL)
}

// ExtractSingleFrame extracts one frame from RTSP stream
func ExtractSingleFrame(rtspURL, outputPath string) error {
	log.Printf("Extracting frame from: %s", maskPassword(rtspURL))

	// Ensure output directory exists
	if err := createOutputDir(outputPath); err != nil {
		return err
	}

	args := []string{
		"-rtsp_transport", "tcp",
		"-i", rtspURL,
		"-vframes", "1",           // Extract only 1 frame
		"-q:v", "2",              // High quality
		"-y",                     // Overwrite output file
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract frame: %w", err)
	}

	log.Printf("Frame extracted successfully: %s", outputPath)
	return nil
}

// ExtractFrameWithConfig extracts frame using config
func ExtractFrameWithConfig(cfg *config.Config, outputPath string) error {
	rtspURL := buildRTSPURL(cfg.RTSPURL, cfg.Username, cfg.Password)
	return ExtractSingleFrame(rtspURL, outputPath)
}

// GenerateThumbnail creates a thumbnail from an image
func GenerateThumbnail(imagePath, outputPath string, size string) error {
	if size == "" {
		size = "320x180" // Default thumbnail size
	}

	log.Printf("Generating thumbnail: %s -> %s (size: %s)", imagePath, outputPath, size)

	// Ensure output directory exists
	if err := createOutputDir(outputPath); err != nil {
		return err
	}

	args := []string{
		"-i", imagePath,
		"-vf", fmt.Sprintf("scale=%s", size),
		"-q:v", "3",              // Good quality for thumbnail
		"-y",                     // Overwrite output file
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	log.Printf("Thumbnail generated: %s", outputPath)
	return nil
}

// RecordVideoSegment records a short video segment
func RecordVideoSegment(rtspURL string, outputPath string, duration int) error {
	log.Printf("Recording %d second segment from: %s", duration, maskPassword(rtspURL))

	// Ensure output directory exists
	if err := createOutputDir(outputPath); err != nil {
		return err
	}

	args := []string{
		"-rtsp_transport", "tcp",
		"-i", rtspURL,
		"-t", fmt.Sprintf("%d", duration), // Record for N seconds
		"-c:v", "libx264",
		"-preset", "fast",
		"-y",                              // Overwrite output file
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to record segment: %w", err)
	}

	log.Printf("Video segment recorded: %s", outputPath)
	return nil
}

// GetVideoInfo extracts basic information about a video file
func GetVideoInfo(videoPath string) (*VideoInfo, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath,
	}

	cmd := exec.Command("ffprobe", args...)
	_, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	// For now, return basic info - you can parse JSON later if needed
	info := &VideoInfo{
		Path:     videoPath,
		Duration: 0, // TODO: parse from JSON output
		Size:     0, // TODO: get file size
	}

	log.Printf("Video info extracted for: %s", videoPath)
	return info, nil
}

// VideoInfo contains basic video information
type VideoInfo struct {
	Path     string
	Duration int   // in seconds
	Size     int64 // in bytes
}

// Utility functions

// buildRTSPURL constructs RTSP URL with credentials
func buildRTSPURL(baseURL, username, password string) string {
	if username == "" || password == "" {
		return baseURL
	}

	// Insert credentials into RTSP URL
	// rtsp://192.168.1.100:554/stream1 -> rtsp://user:pass@192.168.1.100:554/stream1
	if strings.HasPrefix(baseURL, "rtsp://") {
		return fmt.Sprintf("rtsp://%s:%s@%s", username, password, baseURL[7:])
	}
	return baseURL
}

// maskPassword hides password in URL for logging
func maskPassword(url string) string {
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "@")
		if len(parts) >= 2 {
			credPart := parts[0]
			if strings.Contains(credPart, ":") {
				userPass := strings.Split(credPart, ":")
				if len(userPass) >= 2 {
					masked := userPass[0] + ":***"
					return masked + "@" + strings.Join(parts[1:], "@")
				}
			}
		}
	}
	return url
}

// createOutputDir creates directory for output file if it doesn't exist
func createOutputDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return exec.Command("mkdir", "-p", dir).Run()
}

// QuickCameraTest performs a quick test of camera functionality
func QuickCameraTest(cfg *config.Config) error {
	log.Println("Starting quick camera test...")

	// 1. Test connection
	if err := TestRTSPConnectionWithConfig(cfg); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	// 2. Extract test frame
	timestamp := time.Now().Format("20060102_150405")
	framePath := filepath.Join(cfg.OutputDir, fmt.Sprintf("test_frame_%s.jpg", timestamp))
	
	if err := ExtractFrameWithConfig(cfg, framePath); err != nil {
		return fmt.Errorf("frame extraction failed: %w", err)
	}

	// 3. Generate thumbnail
	thumbPath := filepath.Join(cfg.OutputDir, fmt.Sprintf("test_thumb_%s.jpg", timestamp))
	if err := GenerateThumbnail(framePath, thumbPath, "160x90"); err != nil {
		log.Printf("Warning: thumbnail generation failed: %v", err)
		// Don't fail the test for thumbnail error
	}

	log.Println("✅ Quick camera test completed successfully!")
	log.Printf("📸 Test frame: %s", framePath)
	if thumbPath != "" {
		log.Printf("🖼️  Thumbnail: %s", thumbPath)
	}

	return nil
}