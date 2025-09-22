package camera

import (
	"context"
	"fmt"
	"log"
	"time"

	"camera-detection-project/internal/config"
)

// SimpleCamera provides basic camera operations using FFmpeg
type SimpleCamera struct {
	config    *config.Config
	processor *StreamProcessor
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewSimpleCamera(cfg *config.Config) *SimpleCamera {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SimpleCamera{
		config:    cfg,
		processor: NewStreamProcessor(cfg),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Connect tests the connection to the camera
func (sc *SimpleCamera) Connect() error {
	log.Println("Testing camera connection...")
	
	if err := sc.processor.TestConnection(); err != nil {
		return fmt.Errorf("failed to connect to camera: %w", err)
	}
	
	log.Println("Camera connection successful")
	return nil
}

// CaptureFrame captures a single frame from the camera
func (sc *SimpleCamera) CaptureFrame() (string, error) {
	log.Println("Capturing frame...")
	
	framePath, err := sc.processor.ExtractFrame()
	if err != nil {
		return "", fmt.Errorf("failed to capture frame: %w", err)
	}
	
	return framePath, nil
}

// StartCapture starts continuous frame capture
func (sc *SimpleCamera) StartCapture() error {
	log.Println("Starting continuous frame capture...")
	
	ticker := time.NewTicker(time.Duration(sc.config.FrameRate) * time.Second)
	defer ticker.Stop()
	
	frameCount := 0
	
	for {
		select {
		case <-sc.ctx.Done():
			log.Println("Frame capture stopped")
			return nil
		case <-ticker.C:
			frameCount++
			
			framePath, err := sc.CaptureFrame()
			if err != nil {
				log.Printf("Error capturing frame #%d: %v", frameCount, err)
				continue
			}
			
			log.Printf("Frame #%d captured: %s", frameCount, framePath)
			
			// Here you can add processing logic for each frame
			if sc.config.DetectionEnabled {
				sc.processFrame(framePath, frameCount)
			}
		}
	}
}

// processFrame processes a captured frame (placeholder for detection logic)
func (sc *SimpleCamera) processFrame(framePath string, frameNum int) {
	log.Printf("Processing frame #%d: %s", frameNum, framePath)
	
	// Placeholder for object detection
	// You can integrate YOLO, TensorFlow, or other detection libraries here
	
	// Simulate processing time
	time.Sleep(50 * time.Millisecond)
	
	log.Printf("Frame #%d processed", frameNum)
}

// RecordVideo records video for a specified duration
func (sc *SimpleCamera) RecordVideo(duration int) (string, error) {
	log.Printf("Recording video for %d seconds...", duration)
	
	videoPath, err := sc.processor.RecordSegment(duration)
	if err != nil {
		return "", fmt.Errorf("failed to record video: %w", err)
	}
	
	log.Printf("Video recorded: %s", videoPath)
	return videoPath, nil
}

// Stop stops the camera operations
func (sc *SimpleCamera) Stop() {
	log.Println("Stopping simple camera...")
	
	if sc.cancel != nil {
		sc.cancel()
	}
}

// Close closes the camera connection
func (sc *SimpleCamera) Close() error {
	sc.Stop()
	return nil
}