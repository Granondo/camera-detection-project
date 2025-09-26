package camera

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"camera-detection-project/internal/config"
	"camera-detection-project/internal/storage"
)

type FFmpegClient struct {
	config         *config.Config
	cmd            *exec.Cmd
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	frameCount     int
	mu             sync.Mutex
	storageService StorageService
	currentRecording *storage.Recording
}

// StorageService interface to work with storage package
type StorageService interface {
	StartRecording(filePath string) (*storage.Recording, error)
	FinishRecording(recordingID int, filePath string) error
	SaveFrame(filePath string, recordingID *int) (*storage.Frame, error)
	UpdateFrameProcessed(frameID int, hasDetection bool, thumbnailPath *string) error
	CreateEvent(eventType, severity, title, message string, metadata *string) error
	UpdateCameraStatus(status string) error
}

// NewFFmpegClient creates a new FFmpeg client without storage
func NewFFmpegClient(cfg *config.Config) (*FFmpegClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	client := &FFmpegClient{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	return client, nil
}

// NewFFmpegClientWithStorage creates a new FFmpeg client with storage integration
func NewFFmpegClientWithStorage(cfg *config.Config, storage StorageService) (*FFmpegClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	client := &FFmpegClient{
		config:         cfg,
		ctx:            ctx,
		cancel:         cancel,
		storageService: storage,
	}

	return client, nil
}

func (c *FFmpegClient) Start() error {
	log.Println("🎬 Starting FFmpeg video capture...")

	// Create recording record if storage is available
	if c.storageService != nil {
		timestamp := time.Now().Format("20060102_150405")
		recordingPath := filepath.Join(c.config.OutputDir, fmt.Sprintf("recording_%s.mp4", timestamp))
		
		recording, err := c.storageService.StartRecording(recordingPath)
		if err != nil {
			log.Printf("⚠️  Warning: Could not create recording record: %v", err)
		} else {
			c.currentRecording = recording
			log.Printf("📹 Started recording (ID: %d): %s", recording.ID, recordingPath)
		}
	}

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
		go c.watchFrames()
	}

	log.Println("✅ FFmpeg client started successfully")
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

	timestamp := time.Now().Format("20060102_150405")
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
		filepath.Join(c.config.OutputDir, fmt.Sprintf("recording_%s_%%Y%%m%%d_%%H%%M%%S.mp4", timestamp)),
	}

	// Add frame extraction for detection if needed
	if c.config.SaveFrames {
		frameArgs := []string{
			"-vf", fmt.Sprintf("fps=1/%d", c.config.FrameRate), // Extract frame every N seconds
			"-f", "image2",
			"-strftime", "1",
			filepath.Join(c.config.OutputDir, fmt.Sprintf("frame_%s_%%Y%%m%%d_%%H%%M%%S.jpg", timestamp)),
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
			
			// Create error events for important FFmpeg errors
			if c.storageService != nil && name == "STDERR" {
				c.handleFFmpegError(line)
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading %s: %v", name, err)
	}
}

func (c *FFmpegClient) handleFFmpegError(line string) {
	// Check for critical errors and create events
	if contains := []string{"Connection refused", "timeout", "No route to host"}; containsAny(line, contains) {
		c.storageService.CreateEvent(
			"camera_error",
			"high",
			"Camera Connection Error",
			fmt.Sprintf("FFmpeg error: %s", line),
			nil,
		)
	}
}

func (c *FFmpegClient) watchFrames() {
    defer c.wg.Done()
    
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Printf("❌ Failed to create file watcher: %v", err)
        return
    }
    defer watcher.Close()
    
    // Следить за папкой output
    err = watcher.Add(c.config.OutputDir)
    if err != nil {
        log.Printf("❌ Failed to watch directory: %v", err)
        return
    }
    
    log.Printf("👁️ Watching for new frames in: %s", c.config.OutputDir)
    
    for {
        select {
        case <-c.ctx.Done():
            return
        case event, ok := <-watcher.Events:
            if !ok {
                return
            }
            
            // Обработка только создания .jpg файлов
            if event.Op&fsnotify.Create == fsnotify.Create && 
               strings.HasSuffix(event.Name, ".jpg") &&
               strings.Contains(event.Name, "frame_") {
                
                c.handleNewFrame(event.Name)
            }
            
        case err, ok := <-watcher.Errors:
            if !ok {
                return
            }
            log.Printf("⚠️ File watcher error: %v", err)
        }
    }
}

func (c *FFmpegClient) handleNewFrame(framePath string) {
    log.Printf("🖼️ New frame detected: %s", filepath.Base(framePath))
    
    // Небольшая задержка чтобы файл полностью записался
    time.Sleep(100 * time.Millisecond)
    
    // Сохранить в базу данных
    if c.storageService != nil {
        var recordingID *int
        if c.currentRecording != nil {
            recordingID = &c.currentRecording.ID
        }
        
        frame, err := c.storageService.SaveFrame(framePath, recordingID)
        if err != nil {
            log.Printf("⚠️ Warning: Could not save frame to database: %v", err)
            return
        }
        
        log.Printf("💾 Saved frame to database (ID: %d)", frame.ID)
        
        // Запустить детекцию если включена
        if c.config.DetectionEnabled {
            c.mu.Lock()
            c.frameCount++
            frameNum := c.frameCount
            c.mu.Unlock()
            
            hasDetection := c.detectObjects(framePath, frameNum)
            
            // Обновить результаты детекции
            if err := c.storageService.UpdateFrameProcessed(frame.ID, hasDetection, nil); err != nil {
                log.Printf("⚠️ Warning: Could not update frame processed status: %v", err)
            }
        }
    }
}

func (c *FFmpegClient) detectObjects(framePath string, frameNum int) bool {
	// Placeholder for object detection logic
	log.Printf("🔍 Running object detection on frame #%d...", frameNum)
	
	// Simulate detection processing
	time.Sleep(100 * time.Millisecond)
	
	// Simulate random detection (for testing)
	hasDetection := (frameNum%10 == 0) // Every 10th frame has detection
	
	if hasDetection {
		log.Printf("👤 Detection found in frame #%d!", frameNum)
		
		// Create detection event if storage is available
		if c.storageService != nil {
			c.storageService.CreateEvent(
				"person_detected",
				"medium",
				"Person Detected",
				fmt.Sprintf("Person detected in frame #%d at %s", frameNum, framePath),
				nil,
			)
		}
	}
	
	return hasDetection
}

func (c *FFmpegClient) Stop() {
	log.Println("🛑 Stopping FFmpeg client...")
	
	// Finish current recording if storage is available
	if c.storageService != nil && c.currentRecording != nil {
		// In a real implementation, you'd track the actual file path
		recordingPath := filepath.Join(c.config.OutputDir, fmt.Sprintf("recording_%d.mp4", c.currentRecording.ID))
		if err := c.storageService.FinishRecording(c.currentRecording.ID, recordingPath); err != nil {
			log.Printf("⚠️  Warning: Could not finish recording: %v", err)
		} else {
			log.Printf("✅ Finished recording (ID: %d)", c.currentRecording.ID)
		}
	}
	
	if c.cancel != nil {
		c.cancel()
	}
	
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
	
	c.wg.Wait()
	log.Println("✅ FFmpeg client stopped")
}

func (c *FFmpegClient) Close() error {
	c.Stop()
	return nil
}

// Helper functions
func containsAny(str string, substrings []string) bool {
	for _, substring := range substrings {
		if len(str) >= len(substring) {
			for i := 0; i <= len(str)-len(substring); i++ {
				if str[i:i+len(substring)] == substring {
					return true
				}
			}
		}
	}
	return false
}