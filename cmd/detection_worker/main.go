package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"camera-detection-project/internal/analytics"
	"camera-detection-project/internal/config"
	"camera-detection-project/internal/queue"
	"camera-detection-project/internal/storage"
)

func main() {
	log.Println("🔧 Starting Detection Worker...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Initialize PostgreSQL storage
	storageService, err := storage.NewService(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize storage service: %v", err)
	}
	defer storageService.Close()

	// Initialize ClickHouse (опционально)
	var clickhouseClient *analytics.ClickHouseClient
	clickhouseClient, err = analytics.NewClickHouseClient(&analytics.ClickHouseConfig{
		Host:     cfg.ClickHouseHost,
		Port:     cfg.ClickHousePort,
		Database: cfg.ClickHouseDatabase,
		Username: cfg.ClickHouseUsername,
		Password: cfg.ClickHousePassword,
		Debug:    cfg.ClickHouseDebug,
	})

	if err != nil {
		log.Printf("⚠️  Warning: ClickHouse not available: %v", err)
		clickhouseClient = nil
	} else {
		defer clickhouseClient.Close()
		log.Println("✅ Connected to ClickHouse")
	}

	// Initialize RabbitMQ
	queueClient, err := queue.NewRabbitMQClient(&queue.RabbitMQConfig{
		Host:     cfg.RabbitMQHost,
		Port:     cfg.RabbitMQPort,
		User:     cfg.RabbitMQUser,
		Password: cfg.RabbitMQPassword,
		VHost:    cfg.RabbitMQVHost,
	})

	if err != nil {
		log.Fatalf("❌ Failed to connect to RabbitMQ: %v", err)
	}
	defer queueClient.Close()

	// Create worker
	worker := &DetectionWorker{
		storage:    storageService,
		analytics:  clickhouseClient,
		httpClient: &http.Client{Timeout: cfg.DetectionService.Timeout},
		config:     cfg,
	}

	// Start consuming frames
	err = queueClient.Consume(queue.QueueFramesToDetect, worker.ProcessFrame)
	if err != nil {
		log.Fatalf("❌ Failed to start consuming: %v", err)
	}

	log.Println("✅ Detection Worker started successfully")
	log.Println("🔄 Waiting for frames to process...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("🛑 Shutting down worker...")
}

// DetectionWorker обрабатывает кадры из очереди
type DetectionWorker struct {
	storage    *storage.Service
	analytics  *analytics.ClickHouseClient
	httpClient *http.Client
	config     *config.Config
}

// ProcessFrame обрабатывает сообщение с кадром
func (w *DetectionWorker) ProcessFrame(body []byte) error {
	var frameMsg queue.FrameMessage
	if err := json.Unmarshal(body, &frameMsg); err != nil {
		return fmt.Errorf("failed to unmarshal frame message: %w", err)
	}

	log.Printf("🔍 Processing frame #%d (ID: %d) from %s",
		frameMsg.FrameNumber, frameMsg.FrameID, frameMsg.FilePath)

	// Запустить детекцию
	detections, processingTime, err := w.detectObjects(frameMsg.FilePath)
	if err != nil {
		log.Printf("❌ Detection failed for frame %d: %v", frameMsg.FrameID, err)
		return err
	}

	hasDetection := len(detections) > 0

	// Обновить статус в PostgreSQL
	if err := w.storage.UpdateFrameProcessed(frameMsg.FrameID, hasDetection, nil); err != nil {
		log.Printf("⚠️  Failed to update frame status: %v", err)
	}

	// Логировать в ClickHouse
	if hasDetection && w.analytics != nil {
		if err := w.logDetectionsToClickHouse(detections, frameMsg, processingTime); err != nil {
			log.Printf("⚠️  Failed to log to ClickHouse: %v", err)
		}
	}

	log.Printf("✅ Frame #%d processed: %d objects found in %.1fms",
		frameMsg.FrameNumber, len(detections), processingTime)

	return nil
}

// detectObjects вызывает YOLO detection service
func (w *DetectionWorker) detectObjects(imagePath string) ([]queue.DetectionObject, float64, error) {
	// Преобразовать путь для Docker
	detectionPath := imagePath
	if w.config.OutputDir == "./output" {
		detectionPath = "/app/data/" + imagePath[len("./output/"):]
	}

	requestBody := map[string]string{
		"image_path": detectionPath,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, 0, err
	}

	detectURL := w.config.DetectionService.URL + "/detect"
	resp, err := w.httpClient.Post(detectURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("detection service returned status: %d", resp.StatusCode)
	}

	var result struct {
		Success          bool    `json:"success"`
		TotalObjects     int     `json:"total_objects"`
		ProcessingTimeMS float64 `json:"processing_time_ms"`
		Detections       []struct {
			Class      string  `json:"class"`
			ClassID    int     `json:"class_id"`
			Confidence float64 `json:"confidence"`
			BBox       struct {
				X1 float64 `json:"x1"`
				Y1 float64 `json:"y1"`
				X2 float64 `json:"x2"`
				Y2 float64 `json:"y2"`
			} `json:"bbox"`
		} `json:"detections"`
		Error string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	if !result.Success {
		return nil, 0, fmt.Errorf("detection failed: %s", result.Error)
	}

	// Фильтровать по threshold
	var validDetections []queue.DetectionObject
	for _, det := range result.Detections {
		if det.Confidence >= w.config.DetectionService.ConfidenceThreshold {
			validDetections = append(validDetections, queue.DetectionObject{
				Class:      det.Class,
				ClassID:    det.ClassID,
				Confidence: det.Confidence,
				BBox: queue.BBox{
					X1: det.BBox.X1,
					Y1: det.BBox.Y1,
					X2: det.BBox.X2,
					Y2: det.BBox.Y2,
				},
			})
		}
	}

	return validDetections, result.ProcessingTimeMS, nil
}

// logDetectionsToClickHouse логирует детекции в ClickHouse
func (w *DetectionWorker) logDetectionsToClickHouse(
	detections []queue.DetectionObject,
	frameMsg queue.FrameMessage,
	processingTime float64,
) error {
	if w.analytics == nil {
		return nil
	}

	detectionBatch := make([]analytics.Detection, 0, len(detections))

	recordingID := uint64(0)
	if frameMsg.RecordingID != nil {
		recordingID = uint64(*frameMsg.RecordingID)
	}

	for _, det := range detections {
		clickhouseDet := analytics.Detection{
			Timestamp:        frameMsg.Timestamp,
			CameraID:         uint32(frameMsg.CameraID),
			RecordingID:      recordingID,
			FrameID:          uint64(frameMsg.FrameID),
			ObjectClass:      det.Class,
			Confidence:       float32(det.Confidence),
			BBoxX1:           float32(det.BBox.X1),
			BBoxY1:           float32(det.BBox.Y1),
			BBoxX2:           float32(det.BBox.X2),
			BBoxY2:           float32(det.BBox.Y2),
			ModelVersion:     "yolov8n",
			ProcessingTimeMs: uint32(processingTime),
			TrackingID:       nil,
		}
		detectionBatch = append(detectionBatch, clickhouseDet)
	}

	return w.analytics.InsertDetectionsBatch(detectionBatch)
}