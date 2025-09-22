package camera

import (
	"io"
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"camera-detection-project/internal/config"
)

type FFmpegClient struct {
	config     *config.Config
	cmd        *exec.Cmd
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	frameCount int
	mu         sync.Mutex
}

func NewFFmpegClient(cfg *config.Config) (*FFmpegClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	client := &FFmpegClient{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	return client, nil
}

func (c *FFmpegClient) Start() error {
	log.Println("Starting FFmpeg video capture...")

	// Build FFmpeg command for RTSP stream processing
	args := c.buildFFmpegArgs()
	
	c.cmd = exec.CommandContext(c.ctx, c.config.FFmpegPath, args...)
	
	// Setup stdout and stderr pipes
	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start FFmpeg process
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Start monitoring goroutines
	c.wg.Add(2)
	
	go c.monitorOutput(stdout, "STDOUT")
	go c.monitorOutput(stderr, "STDERR")

	// Start frame processing if detection is enabled
	if c.config.DetectionEnabled {
		c.wg.Add(1)
		go c.processFrames()
	}

	log.Println("FFmpeg client started successfully")
	return nil
}

func (c *FFmpegClient) buildFFmpegArgs() []string {
	// Build RTSP URL with credentials if provided
	rtspURL := c.config.RTSPURL
	if c.config.Username != "" && c.config.Password != "" {
		// Insert credentials into RTSP URL
		rtspURL = fmt.Sprintf("rtsp://%s:%s@%s", 
			c.config.Username, 
			c.config.Password, 
			c.config.RTSPURL[7:]) // Remove "rtsp://" prefix
	}

	args := []string{
		"-rtsp_transport", "tcp",  // Use TCP for RTSP (more reliable)
		"-i", rtspURL,
		"-c:v", "libx264",         // Video codec
		"-preset", "ultrafast",    // Encoding speed
		"-tune", "zerolatency",    // Low latency
		"-f", "segment",           // Output format
		"-segment_time", "60",     // 60 second segments
		"-segment_format", "mp4",  // Segment format
		"-strftime", "1",          // Enable strftime in filename
		filepath.Join(c.config.OutputDir, "recording_%Y%m%d_%H%M%S.mp4"),
	}

	// Add frame extraction for detection if needed
	if c.config.SaveFrames {
		frameArgs := []string{
			"-vf", fmt.Sprintf("fps=1/%d", c.config.FrameRate), // Extract frame every N seconds
			"-f", "image2",
			"-strftime", "1",
			filepath.Join(c.config.OutputDir, "frame_%Y%m%d_%H%M%S.jpg"),
		}
		args = append(args, frameArgs...)
	}

	return args
}

func (c *FFmpegClient) monitorOutput(pipe io.ReadCloser, name string) {
	defer c.wg.Done()
	defer pipe.Close()

	
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		select {
		case <-c.ctx.Done():
			return
		default:
			line := scanner.Text()
			log.Printf("[FFmpeg %s]: %s", name, line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading %s: %v", name, err)
	}
}

func (c *FFmpegClient) processFrames() {
	defer c.wg.Done()
	
	ticker := time.NewTicker(time.Duration(c.config.FrameRate) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			c.frameCount++
			frameNum := c.frameCount
			c.mu.Unlock()
			
			log.Printf("Processing frame #%d", frameNum)
			
			// Here you can add detection logic
			// For now, just log that we're processing
			if c.config.DetectionEnabled {
				c.detectObjects()
			}
		}
	}
}

func (c *FFmpegClient) detectObjects() {
	// Placeholder for object detection logic
	// This can be extended with YOLO, TensorFlow, or other detection models
	log.Println("Running object detection...")
	
	// Example: You could call external detection tools or APIs here
	// For now, just simulate detection
	time.Sleep(100 * time.Millisecond)
}

func (c *FFmpegClient) Stop() {
	log.Println("Stopping FFmpeg client...")
	
	if c.cancel != nil {
		c.cancel()
	}
	
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
	
	c.wg.Wait()
	log.Println("FFmpeg client stopped")
}

func (c *FFmpegClient) Close() error {
	c.Stop()
	return nil
}