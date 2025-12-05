package camera

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"camera-detection-project/internal/analytics"
	"camera-detection-project/internal/config"
	"camera-detection-project/internal/queue"
	"camera-detection-project/internal/storage"
	"github.com/fsnotify/fsnotify"
)

type DetectionResponse struct {
	Success          bool        `json:"success"`
	ImagePath        string      `json:"image_path"`
	Detections       []Detection `json:"detections"`
	TotalObjects     int         `json:"total_objects"`
	ProcessingTimeMS float64     `json:"processing_time_ms"`
	Error            string      `json:"error,omitempty"`
}

type Detection struct {
	Class      string      `json:"class"`
	ClassID    int         `json:"class_id"`
	Confidence float64     `json:"confidence"`
	BBox       BoundingBox `json:"bbox"`
}

type BoundingBox struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

type FFmpegClient struct {
	config             *config.Config
	cmd                *exec.Cmd
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
	frameCount         int
	mu                 sync.Mutex
	storageService     StorageService
	currentRecording   *storage.Recording
	currentRecordingID *int // ← ДОБАВЛЕНО для tracking
	detectionClient    *http.Client
	analyticsClient    *analytics.ClickHouseClient
	queueClient        *queue.RabbitMQClient
	searchClient       SearchClient // ← NEW: Elasticsearch client
	cameraID           int // ← ДОБАВЛЕНО для ClickHouse
	recordingDetections map[int]bool // ← Tracks if recording has any detections
}

// StorageService interface to work with storage package
type StorageService interface {
	StartRecording(filePath string) (*storage.Recording, error)
	FinishRecording(recordingID int, filePath string) error
	SaveFrame(filePath string, recordingID *int) (*storage.Frame, error)
	UpdateFrameProcessed(frameID int, hasDetection bool, thumbnailPath *string) error
	CreateEvent(eventType, severity, title, message string, metadata *string) (*storage.Event, error)
	CreateEventWithFrame(eventType, severity, title, message string, metadata *string, frameID *int) (*storage.Event, error)
	UpdateCameraStatus(status string) error
	GetCameraStatus() (*storage.Camera, error) // ← ДОБАВЛЕНО
	GetRecording(recordingID int) (*storage.Recording, error) // ← ДОБАВЛЕНО для cleanup
	DeleteRecording(recordingID int) error // ← ДОБАВЛЕНО для cleanup
}

// SearchClient interface for Elasticsearch
type SearchClient interface {
	IndexDocument(ctx context.Context, doc interface{}) error
}

// NewFFmpegClient creates a new FFmpeg client without storage
func NewFFmpegClient(cfg *config.Config) (*FFmpegClient, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &FFmpegClient{
		config:              cfg,
		ctx:                 ctx,
		cancel:              cancel,
		cameraID:            1, // default
		recordingDetections: make(map[int]bool),
	}

	return client, nil
}

// NewFFmpegClientWithStorage creates a new FFmpeg client with storage integration
func NewFFmpegClientWithStorage(cfg *config.Config, storage StorageService) (*FFmpegClient, error) {
	ctx, cancel := context.WithCancel(context.Background())

	detectionClient := &http.Client{
		Timeout: cfg.DetectionService.Timeout,
	}

	client := &FFmpegClient{
		config:              cfg,
		ctx:                 ctx,
		cancel:              cancel,
		storageService:      storage,
		detectionClient:     detectionClient,
		cameraID:            1, // default
		recordingDetections: make(map[int]bool),
	}

	return client, nil
}

// NewFFmpegClientWithStorageAndAnalytics creates client with storage and analytics
func NewFFmpegClientWithStorageAndAnalytics(
	cfg *config.Config,
	storage StorageService,
	analytics *analytics.ClickHouseClient,
	queueClient *queue.RabbitMQClient,
	searchClient SearchClient,
) (*FFmpegClient, error) {
	ctx, cancel := context.WithCancel(context.Background())

	detectionClient := &http.Client{
		Timeout: cfg.DetectionService.Timeout,
	}

	client := &FFmpegClient{
		config:              cfg,
		ctx:                 ctx,
		cancel:              cancel,
		storageService:      storage,
		detectionClient:     detectionClient,
		analyticsClient:     analytics,
		queueClient:         queueClient,
		searchClient:        searchClient,
		cameraID:            1, // default
		recordingDetections: make(map[int]bool),
	}

	// Получить реальный ID камеры из БД
	if storage != nil {
		if camera, err := storage.GetCameraStatus(); err == nil && camera != nil {
			client.cameraID = camera.ID
			log.Printf("📹 Using camera ID: %d", client.cameraID)
		}
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
			c.currentRecordingID = &recording.ID
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

	c.wg.Add(1)
	go c.watchFrames()

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
		"-rtsp_transport", "tcp", // Use TCP for RTSP (more reliable)
		"-i", rtspURL,
		"-c:v", "libx264", // Video codec
		"-preset", "ultrafast", // Encoding speed
		"-tune", "zerolatency", // Low latency
		"-f", "segment", // Output format
		"-segment_time", "60", // 60 second segments
		"-segment_format", "mp4", // Segment format
		"-strftime", "1", // Enable strftime in filename
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

			if event.Op&fsnotify.Create == fsnotify.Create &&
				strings.HasSuffix(event.Name, ".mp4") &&
				strings.Contains(event.Name, "recording_") {

				c.handleNewRecording(event.Name)
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

	c.mu.Lock()
	c.frameCount++
	frameNum := c.frameCount
	c.mu.Unlock()

	// ✅ NEW APPROACH: Run detection FIRST, before saving
	if !c.config.DetectionEnabled {
		log.Printf("⏭️  Detection disabled, skipping frame #%d", frameNum)
		// Delete frame since we're not processing it
		if err := os.Remove(framePath); err != nil {
			log.Printf("⚠️  Could not delete frame: %v", err)
		}
		return
	}

	log.Printf("🔍 Running detection on frame #%d BEFORE saving...", frameNum)
	hasDetection, detections, processingTime := c.detectObjects(framePath, frameNum)

	// ✅ Only save frame if detection found
	if !hasDetection {
		log.Printf("❌ No detection in frame #%d - deleting frame", frameNum)
		if err := os.Remove(framePath); err != nil {
			log.Printf("⚠️  Could not delete frame: %v", err)
		}
		return
	}

	// ✅ Detection found! Save to database
	log.Printf("✅ Detection found in frame #%d - saving to database", frameNum)

	if c.storageService == nil {
		log.Printf("⚠️  No storage service, cannot save frame")
		return
	}

	frame, err := c.storageService.SaveFrame(framePath, c.currentRecordingID)
	if err != nil {
		log.Printf("⚠️  Warning: Could not save frame to database: %v", err)
		return
	}

	log.Printf("💾 Saved frame to database (ID: %d)", frame.ID)

	// Mark frame as processed with detection
	if err := c.storageService.UpdateFrameProcessed(frame.ID, true, nil); err != nil {
		log.Printf("⚠️  Warning: Could not update frame processed status: %v", err)
	}

	// Track that this recording has a detection
	c.mu.Lock()
	if c.currentRecordingID != nil {
		c.recordingDetections[*c.currentRecordingID] = true
	}
	c.mu.Unlock()

	// Create detection event with frame ID
	if len(detections) > 0 && c.storageService != nil {
		c.createDetectionEventWithFrame(detections, framePath, frameNum, frame.ID)
	}

	// Log detections to ClickHouse
	if len(detections) > 0 && c.analyticsClient != nil {
		c.logDetectionsToClickHouse(detections, frame.ID, frameNum, processingTime)
	}
}

// processFrameSync синхронная обработка кадра (старый способ без RabbitMQ)
func (c *FFmpegClient) processFrameSync(frame *storage.Frame, framePath string, frameNum int) {
	if c.config.DetectionEnabled {
		// Запустить детекцию (HTTP вызов к detection service)
		hasDetection, detections, processingTime := c.detectObjects(framePath, frameNum)

		// Обновить результаты детекции в PostgreSQL
		if err := c.storageService.UpdateFrameProcessed(frame.ID, hasDetection, nil); err != nil {
			log.Printf("⚠️ Warning: Could not update frame processed status: %v", err)
		}

		// Логирование в ClickHouse (если есть детекции)
		if hasDetection && len(detections) > 0 && c.analyticsClient != nil {
			c.logDetectionsToClickHouse(detections, frame.ID, frameNum, processingTime)
		}
	}
}

func (c *FFmpegClient) handleNewRecording(videoPath string) {
	log.Printf("🎥 New video segment detected: %s", filepath.Base(videoPath))

	// Подождать чтобы файл полностью записался
	time.Sleep(1 * time.Second)

	// Check if previous recording had any detections
	c.mu.Lock()
	previousRecordingID := c.currentRecordingID
	c.mu.Unlock()

	if previousRecordingID != nil {
		c.mu.Lock()
		hadDetections := c.recordingDetections[*previousRecordingID]
		c.mu.Unlock()

		if !hadDetections {
			// No detections in previous recording - find and delete it
			log.Printf("🗑️  Previous recording (ID: %d) had no detections - will be deleted", *previousRecordingID)
			c.cleanupRecordingWithoutDetections(*previousRecordingID)
		} else {
			log.Printf("✅ Previous recording (ID: %d) has detections - keeping", *previousRecordingID)
		}
	}

	// Получить информацию о файле
	fileInfo, err := os.Stat(videoPath)
	if err != nil {
		log.Printf("⚠️ Could not get video file info: %v", err)
		return
	}

	// Создать запись в БД для нового сегмента
	if c.storageService != nil {
		recording, err := c.storageService.StartRecording(videoPath)
		if err != nil {
			log.Printf("⚠️ Could not create recording record: %v", err)
			return
		}

		// ✅ Обновить текущий recording ID
		c.mu.Lock()
		c.currentRecording = recording
		c.currentRecordingID = &recording.ID
		// Initialize detection tracking for new recording
		c.recordingDetections[recording.ID] = false
		c.mu.Unlock()

		// Сразу финализировать запись (т.к. сегмент уже завершен)
		if err := c.storageService.FinishRecording(recording.ID, videoPath); err != nil {
			log.Printf("⚠️ Could not finish recording: %v", err)
		} else {
			log.Printf("✅ Saved recording to database (ID: %d, Size: %s)",
				recording.ID, formatBytes(fileInfo.Size()))
		}
	}
}

// cleanupRecordingWithoutDetections deletes recording file and DB record
func (c *FFmpegClient) cleanupRecordingWithoutDetections(recordingID int) {
	if c.storageService == nil {
		return
	}

	// Get recording info from database
	recording, err := c.storageService.GetRecording(recordingID)
	if err != nil {
		log.Printf("⚠️  Could not get recording %d for cleanup: %v", recordingID, err)
		return
	}

	// Delete the video file from disk
	if err := os.Remove(recording.FilePath); err != nil {
		log.Printf("⚠️  Could not delete recording file %s: %v", recording.FilePath, err)
	} else {
		log.Printf("🗑️  Deleted recording file: %s", filepath.Base(recording.FilePath))
	}

	// Delete from database
	if err := c.storageService.DeleteRecording(recordingID); err != nil {
		log.Printf("⚠️  Could not delete recording from database: %v", err)
	} else {
		log.Printf("✅ Deleted recording from database (ID: %d)", recordingID)
	}

	// Clean up from tracking map
	c.mu.Lock()
	delete(c.recordingDetections, recordingID)
	c.mu.Unlock()
}

// ✅ НОВЫЙ МЕТОД: Логирование детекций в ClickHouse
func (c *FFmpegClient) logDetectionsToClickHouse(detections []Detection, frameID int, frameNum int, processingTimeMs float64) {
	if c.analyticsClient == nil {
		return
	}

	detectionBatch := make([]analytics.Detection, 0, len(detections))

	for _, det := range detections {
		// Получить recording ID (безопасно)
		var recordingID uint64 = 0
		c.mu.Lock()
		if c.currentRecordingID != nil {
			recordingID = uint64(*c.currentRecordingID)
		}
		c.mu.Unlock()

		clickhouseDet := analytics.Detection{
			Timestamp:        time.Now(),
			CameraID:         uint32(c.cameraID),
			RecordingID:      recordingID,
			FrameID:          uint64(frameID),
			ObjectClass:      det.Class,
			Confidence:       float32(det.Confidence),
			BBoxX1:           float32(det.BBox.X1),
			BBoxY1:           float32(det.BBox.Y1),
			BBoxX2:           float32(det.BBox.X2),
			BBoxY2:           float32(det.BBox.Y2),
			ModelVersion:     "yolov8n",
			ProcessingTimeMs: uint32(processingTimeMs), // ← ИСПОЛЬЗУЕМ ПАРАМЕТР
			TrackingID:       nil,
		}
		detectionBatch = append(detectionBatch, clickhouseDet)
	}

	// Batch insert в ClickHouse
	if err := c.analyticsClient.InsertDetectionsBatch(detectionBatch); err != nil {
		log.Printf("⚠️  Failed to log detections to ClickHouse: %v", err)
	} else {
		log.Printf("✅ Logged %d detections to ClickHouse (Frame #%d, %.1fms)", len(detections), frameNum, processingTimeMs)
	}
}

func (c *FFmpegClient) detectObjects(framePath string, frameNum int) (bool, []Detection, float64) {
	if !c.config.DetectionEnabled {
		return false, nil, 0
	}

	log.Printf("🔍 Running YOLO detection on frame #%d: %s", frameNum, filepath.Base(framePath))

	detectionPath := strings.Replace(framePath, c.config.OutputDir, "/app/data", 1)

	log.Printf("🔄 Transformed path: %s -> %s", framePath, detectionPath)

	// Подготовить запрос
	requestBody := map[string]string{
		"image_path": detectionPath,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("❌ Failed to marshal detection request: %v", err)
		return false, nil, 0
	}

	// Отправить запрос к detection service с retry логикой
	var result DetectionResponse
	var lastErr error

	for attempt := 1; attempt <= c.config.DetectionService.MaxRetries; attempt++ {
		detectURL := c.config.DetectionService.URL + "/detect"
		resp, err := c.detectionClient.Post(detectURL, "application/json", bytes.NewBuffer(jsonData))

		if err != nil {
			lastErr = err
			log.Printf("⚠️  Detection attempt %d/%d failed: %v", attempt, c.config.DetectionService.MaxRetries, err)
			if attempt < c.config.DetectionService.MaxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			break
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("detection service returned status: %d", resp.StatusCode)
			log.Printf("⚠️  Detection attempt %d/%d failed with status: %d", attempt, c.config.DetectionService.MaxRetries, resp.StatusCode)
			if attempt < c.config.DetectionService.MaxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			break
		}

		// Разобрать ответ
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			lastErr = err
			log.Printf("⚠️  Detection attempt %d/%d failed to decode response: %v", attempt, c.config.DetectionService.MaxRetries, err)
			if attempt < c.config.DetectionService.MaxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			break
		}

		// Успешно получили ответ
		lastErr = nil
		break
	}

	if lastErr != nil {
		log.Printf("❌ All detection attempts failed: %v", lastErr)
		c.logDetectionError(lastErr.Error())
		return false, nil, 0
	}

	if !result.Success {
		log.Printf("❌ Detection failed: %s", result.Error)
		c.logDetectionError(result.Error)
		return false, nil, 0
	}

	// Обработать результаты
	if result.TotalObjects > 0 {
		log.Printf("✅ Found %d objects in %.1fms:", result.TotalObjects, result.ProcessingTimeMS)

		// Фильтровать по порогу уверенности
		validDetections := []Detection{}
		for _, detection := range result.Detections {
			if detection.Confidence >= c.config.DetectionService.ConfidenceThreshold {
				validDetections = append(validDetections, detection)
				confidence := detection.Confidence * 100
				log.Printf("   🎯 %s (%.1f%%)", detection.Class, confidence)
			}
		}

		if len(validDetections) > 0 {
			// Event creation will happen after frame is saved in handleNewFrame
			return true, validDetections, result.ProcessingTimeMS // ← ВОЗВРАЩАЕМ ДЕТЕКЦИИ
		} else {
			log.Printf("📷 Objects found but below confidence threshold (%.2f) in frame #%d",
				c.config.DetectionService.ConfidenceThreshold, frameNum)
			return false, nil, 0
		}
	} else {
		log.Printf("📷 No objects detected in frame #%d (%.1fms)", frameNum, result.ProcessingTimeMS)
		return false, nil, 0
	}
}

func (c *FFmpegClient) createDetectionEventWithFrame(detections []Detection, framePath string, frameNum int, frameID int) {
	var mainDetection Detection
	var maxConfidence float64 = 0

	for _, det := range detections {
		if det.Confidence > maxConfidence {
			maxConfidence = det.Confidence
			mainDetection = det
		}
	}

	if maxConfidence == 0 {
		return
	}

	// Определить серьезность события
	severity := "low"
	if maxConfidence > 0.7 {
		severity = "medium"
	}
	if maxConfidence > 0.9 {
		severity = "high"
	}

	// Определить тип события в зависимости от класса объекта
	eventType := "object_detected"
	if mainDetection.Class == "person" {
		eventType = "person_detected"
	} else if mainDetection.Class == "car" || mainDetection.Class == "truck" {
		eventType = "vehicle_detected"
	}

	title := fmt.Sprintf("%s Detected", strings.Title(mainDetection.Class))
	message := fmt.Sprintf("%s detected in frame #%d with %.1f%% confidence at %s",
		strings.Title(mainDetection.Class), frameNum, maxConfidence*100, filepath.Base(framePath))

	if len(detections) > 1 {
		message += fmt.Sprintf(" (total: %d objects)", len(detections))
	}

	// Создать событие в PostgreSQL с привязкой к frame
	event, err := c.storageService.CreateEventWithFrame(
		eventType,
		severity,
		title,
		message,
		nil,
		&frameID,
	)

	if err != nil {
		log.Printf("⚠️  Failed to create detection event: %v", err)
		return
	}

	// ✅ NEW: Index to Elasticsearch asynchronously
	if c.searchClient != nil && event != nil {
		go c.indexEventToElasticsearch(event, framePath)
	}

	// Логировать в ClickHouse
	if c.analyticsClient != nil {

		cameraIDUint32 := uint32(c.cameraID)

		event := &analytics.SystemEvent{
			Timestamp:   time.Now(),
			EventType:   eventType,
			Severity:    severity,
			Service:     "detection-service",
			Title:       title,
			Message:     message,
			CameraID:    &cameraIDUint32,
			RecordingID: nil,
			FrameID:     nil,
			Metadata:    fmt.Sprintf(`{"frame_number": %d, "detections_count": %d, "max_confidence": %.2f}`, frameNum, len(detections), maxConfidence),
		}

		c.mu.Lock()
		if c.currentRecordingID != nil {
			recordingID := uint64(*c.currentRecordingID)
			event.RecordingID = &recordingID
		}
		c.mu.Unlock()

		if err := c.analyticsClient.InsertSystemEvent(event); err != nil {
			log.Printf("⚠️  Failed to log event to ClickHouse: %v", err)
		}
	}
}

func (c *FFmpegClient) logDetectionError(errorMsg string) {
	if c.storageService != nil {
		event, err := c.storageService.CreateEvent(
			"detection_error",
			"medium",
			"Detection Service Error",
			fmt.Sprintf("Detection service failed: %s", errorMsg),
			nil,
		)

		// Index to Elasticsearch
		if err == nil && c.searchClient != nil && event != nil {
			go c.indexEventToElasticsearch(event, "")
		}
	}

	// Логировать в ClickHouse
	if c.analyticsClient != nil {
		cameraIDUint32 := uint32(c.cameraID)

		event := &analytics.SystemEvent{
			Timestamp:   time.Now(),
			EventType:   "detection_error",
			Severity:    "medium",
			Service:     "detection-service",
			Title:       "Detection Service Error",
			Message:     fmt.Sprintf("Detection service failed: %s", errorMsg),
			CameraID:    &cameraIDUint32,
			RecordingID: nil,
			FrameID:     nil,
			Metadata:    fmt.Sprintf(`{"error": "%s"}`, errorMsg),
		}

		if err := c.analyticsClient.InsertSystemEvent(event); err != nil {
			log.Printf("⚠️  Failed to log error to ClickHouse: %v", err)
		}
	}
}

// indexEventToElasticsearch indexes an event to Elasticsearch
func (c *FFmpegClient) indexEventToElasticsearch(event *storage.Event, framePath string) {
	if c.searchClient == nil || event == nil {
		return
	}

	// Create search document
	doc := map[string]interface{}{
		"event_id":   event.ID,
		"frame_id":   event.FrameID,
		"camera_id":  event.CameraID,
		"timestamp":  event.Timestamp,
		"event_type": event.EventType,
		"severity":   event.Severity,
		"title":      event.Title,
		"message":    event.Message,
		"has_frame":  event.FrameID != nil,
	}

	if framePath != "" {
		doc["frame_path"] = framePath
	}

	// Index document
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.searchClient.IndexDocument(ctx, doc); err != nil {
		log.Printf("⚠️  Failed to index event to Elasticsearch: %v", err)
	} else {
		log.Printf("✅ Indexed event #%d to Elasticsearch", event.ID)
	}
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
