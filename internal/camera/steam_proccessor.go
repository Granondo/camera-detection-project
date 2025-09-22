package camera

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"camera-detection-project/internal/config"
)

type StreamProcessor struct {
	config *config.Config
}

func NewStreamProcessor(cfg *config.Config) *StreamProcessor {
	return &StreamProcessor{
		config: cfg,
	}
}

// ExtractFrame extracts a single frame from RTSP stream
func (sp *StreamProcessor) ExtractFrame() (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	outputPath := filepath.Join(sp.config.OutputDir, fmt.Sprintf("frame_%s.jpg", timestamp))
	
	rtspURL := sp.config.RTSPURL
	if sp.config.Username != "" && sp.config.Password != "" {
		rtspURL = fmt.Sprintf("rtsp://%s:%s@%s", 
			sp.config.Username, 
			sp.config.Password, 
			sp.config.RTSPURL[7:])
	}

	args := []string{
		"-rtsp_transport", "tcp",
		"-i", rtspURL,
		"-vframes", "1",           // Extract only 1 frame
		"-q:v", "2",              // High quality
		"-y",                     // Overwrite output file
		outputPath,
	}

	cmd := exec.Command(sp.config.FFmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to extract frame: %w", err)
	}

	log.Printf("Frame extracted: %s", outputPath)
	return outputPath, nil
}

// TestConnection tests RTSP connection using ffprobe
func (sp *StreamProcessor) TestConnection() error {
	rtspURL := sp.config.RTSPURL
	if sp.config.Username != "" && sp.config.Password != "" {
		rtspURL = fmt.Sprintf("rtsp://%s:%s@%s", 
			sp.config.Username, 
			sp.config.Password, 
			sp.config.RTSPURL[7:])
	}

	args := []string{
		"-rtsp_transport", "tcp",
		"-i", rtspURL,
		"-t", "1",                // Test for 1 second
		"-f", "null",             // No output
		"-",
	}

	cmd := exec.Command("ffprobe", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("RTSP connection test failed: %w", err)
	}

	log.Println("RTSP connection test successful")
	return nil
}

// RecordSegment records a video segment for specified duration
func (sp *StreamProcessor) RecordSegment(duration int) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	outputPath := filepath.Join(sp.config.OutputDir, fmt.Sprintf("segment_%s.mp4", timestamp))
	
	rtspURL := sp.config.RTSPURL
	if sp.config.Username != "" && sp.config.Password != "" {
		rtspURL = fmt.Sprintf("rtsp://%s:%s@%s", 
			sp.config.Username, 
			sp.config.Password, 
			sp.config.RTSPURL[7:])
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

	cmd := exec.Command(sp.config.FFmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to record segment: %w", err)
	}

	log.Printf("Video segment recorded: %s", outputPath)
	return outputPath, nil
}