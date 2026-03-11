package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"camera-detection-project/internal/analytics"
	"camera-detection-project/internal/config"
)

// Service provides high-level storage operations
type Service struct {
	db              *Database
	cameraRepo      *CameraRepository
	recordingRepo   *RecordingRepository
	frameRepo       *FrameRepository
	eventRepo       *EventRepository
	config          *config.Config
	defaultCameraID int // ID of the main camera
}

func (s *Service) ensureDefaultCamera() error {
	camera, err := s.cameraRepo.GetCamera(s.defaultCameraID)
	if err == nil && camera != nil {
		return nil
	}

	cameras, err := s.cameraRepo.GetAllCameras()
	if err != nil {
		return err
	}
	if len(cameras) > 0 {
		s.defaultCameraID = cameras[0].ID
		return nil
	}

	cameraToCreate := &Camera{
		Name:     "Main Camera",
		RTSPURL:  s.config.RTSPURL,
		Username: s.config.Username,
		Password: s.config.Password,
		Status:   CameraStatusActive,
	}
	if err := s.cameraRepo.CreateCamera(cameraToCreate); err != nil {
		return fmt.Errorf("failed to create default camera: %w", err)
	}
	s.defaultCameraID = cameraToCreate.ID
	log.Printf("Created default camera with ID: %d", cameraToCreate.ID)
	return nil
}

// NewService creates a new storage service
func NewService(cfg *config.Config) (*Service, error) {
	// Create database configuration
	dbConfig := &DatabaseConfig{
		Host:         cfg.DatabaseHost,
		Port:         cfg.DatabasePort,
		User:         cfg.DatabaseUser,
		Password:     cfg.DatabasePassword,
		Database:     cfg.DatabaseName,
		SSLMode:      cfg.DatabaseSSLMode,
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  30 * time.Minute,
	}

	// Connect to database
	db, err := NewDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create repositories
	cameraRepo := NewCameraRepository(db)
	recordingRepo := NewRecordingRepository(db)
	frameRepo := NewFrameRepository(db)
	eventRepo := NewEventRepository(db)

	service := &Service{
		db:            db,
		cameraRepo:    cameraRepo,
		recordingRepo: recordingRepo,
		frameRepo:     frameRepo,
		eventRepo:     eventRepo,
		config:        cfg,
	}

	// Initialize default camera if needed
	if err := service.initializeDefaultCamera(); err != nil {
		return nil, fmt.Errorf("failed to initialize camera: %w", err)
	}

	log.Println("Storage service initialized successfully")
	return service, nil
}

// Close closes the storage service
func (s *Service) Close() error {
	return s.db.Close()
}

// InitializeDatabase creates tables if they don't exist
func (s *Service) InitializeDatabase() error {
	return s.db.CreateTables()
}

// initializeDefaultCamera creates or updates the default camera
func (s *Service) initializeDefaultCamera() error {
	// Try to get existing camera
	cameras, err := s.cameraRepo.GetAllCameras()
	if err != nil {
		return err
	}

	if len(cameras) == 0 {
		// Create default camera
		camera := &Camera{
			Name:     "Main Camera",
			RTSPURL:  s.config.RTSPURL,
			Username: s.config.Username,
			Password: s.config.Password,
			Status:   CameraStatusActive,
		}

		if err := s.cameraRepo.CreateCamera(camera); err != nil {
			return fmt.Errorf("failed to create default camera: %w", err)
		}

		s.defaultCameraID = camera.ID
		log.Printf("Created default camera with ID: %d", camera.ID)
	} else {
		// Use first camera as default
		s.defaultCameraID = cameras[0].ID
		log.Printf("Using existing camera with ID: %d", s.defaultCameraID)
	}

	return nil
}

// Recording methods

// StartRecording creates a new recording record
func (s *Service) StartRecording(filePath string) (*Recording, error) {
	if err := s.ensureDefaultCamera(); err != nil {
		return nil, err
	}

	recording := &Recording{
		CameraID:  s.defaultCameraID,
		FilePath:  filePath,
		StartTime: time.Now(),
		Status:    RecordingStatusRecording,
		Quality:   "1080p", // TODO: detect from stream
		Codec:     "h264",
	}

	if err := s.recordingRepo.CreateRecording(recording); err != nil {
		return nil, fmt.Errorf("failed to create recording: %w", err)
	}

	log.Printf("Started recording: %s (ID: %d)", filePath, recording.ID)
	return recording, nil
}

// FinishRecording updates recording with final information
func (s *Service) FinishRecording(recordingID int, filePath string) error {
	recording, err := s.recordingRepo.GetRecording(recordingID)
	if err != nil {
		return fmt.Errorf("failed to get recording: %w", err)
	}

	// Get file information
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("Warning: could not get file info for %s: %v", filePath, err)
		recording.FileSize = 0
	} else {
		recording.FileSize = fileInfo.Size()
	}

	// Calculate duration
	now := time.Now()
	recording.EndTime = &now
	recording.Duration = int(now.Sub(recording.StartTime).Seconds())
	recording.Status = RecordingStatusCompleted

	if err := s.recordingRepo.UpdateRecording(recording); err != nil {
		return fmt.Errorf("failed to update recording: %w", err)
	}

	log.Printf("Finished recording: %s (Duration: %ds, Size: %d bytes)",
		filePath, recording.Duration, recording.FileSize)

	return nil
}

// Frame methods

// SaveFrame creates a new frame record
func (s *Service) SaveFrame(filePath string, recordingID *int) (*Frame, error) {
	if err := s.ensureDefaultCamera(); err != nil {
		return nil, err
	}

	// Get file information
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	frame := &Frame{
		RecordingID: recordingID,
		CameraID:    s.defaultCameraID,
		FilePath:    filePath,
		FileSize:    int(fileInfo.Size()),
		Timestamp:   time.Now(),
		Width:       1920, // TODO: detect from image
		Height:      1080, // TODO: detect from image
	}

	if err := s.frameRepo.CreateFrame(frame); err != nil {
		return nil, fmt.Errorf("failed to create frame: %w", err)
	}

	log.Printf("Saved frame: %s (ID: %d)", filePath, frame.ID)
	return frame, nil
}

// UpdateFrameProcessed marks frame as processed with detection results
func (s *Service) UpdateFrameProcessed(frameID int, hasDetection bool, thumbnailPath *string) error {
	frame, err := s.frameRepo.GetFrame(frameID)
	if err != nil {
		return fmt.Errorf("failed to get frame: %w", err)
	}

	frame.HasDetection = hasDetection
	frame.Processed = true
	frame.ThumbnailPath = thumbnailPath

	if err := s.frameRepo.UpdateFrame(frame); err != nil {
		return fmt.Errorf("failed to update frame: %w", err)
	}

	return nil
}

// GetUnprocessedFrames retrieves frames that need processing
func (s *Service) GetUnprocessedFrames(limit int) ([]Frame, error) {
	return s.frameRepo.GetUnprocessedFrames(limit)
}

// Event methods

// CreateEvent creates a new event
func (s *Service) CreateEvent(eventType, severity, title, message string, metadata *string) (*Event, error) {
	return s.CreateEventWithFrame(eventType, severity, title, message, metadata, nil)
}

// CreateEventWithFrame creates a new event linked to a specific frame
func (s *Service) CreateEventWithFrame(eventType, severity, title, message string, metadata *string, frameID *int) (*Event, error) {
	event := &Event{
		CameraID:  &s.defaultCameraID,
		FrameID:   frameID,
		EventType: eventType,
		Severity:  severity,
		Title:     title,
		Message:   message,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	if err := s.eventRepo.CreateEvent(event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	if frameID != nil {
		log.Printf("Created event: %s (%s) - %s (Frame ID: %d)", eventType, severity, title, *frameID)
	} else {
		log.Printf("Created event: %s (%s) - %s", eventType, severity, title)
	}
	return event, nil
}

// CreateEventWithAnalytics создает событие в PostgreSQL и логирует в ClickHouse
func (s *Service) CreateEventWithAnalytics(eventType, severity, title, message string, metadata *string, clickhouseClient interface{}) (*Event, error) {
	// Сначала сохраняем в PostgreSQL
	event, err := s.CreateEvent(eventType, severity, title, message, metadata)
	if err != nil {
		return nil, err
	}

	// Потом логируем в ClickHouse если доступен
	if clickhouseClient != nil {

		cameraIDUint32 := uint32(s.defaultCameraID)
		if chClient, ok := clickhouseClient.(ClickHouseLogger); ok {
			event := &analytics.SystemEvent{
				Timestamp:   time.Now(),
				EventType:   eventType,
				Severity:    severity,
				Service:     "camera-service",
				Title:       title,
				Message:     message,
				CameraID:    &cameraIDUint32,
				RecordingID: nil,
				FrameID:     nil,
				Metadata:    "{}",
			}
			if metadata != nil {
				event.Metadata = *metadata
			}

			if err := chClient.InsertSystemEvent(event); err != nil {
				log.Printf("⚠️  Warning: Could not log event to ClickHouse: %v", err)
			}
		}
	}

	return event, nil
}

// ClickHouseLogger интерфейс для логирования в ClickHouse
type ClickHouseLogger interface {
	InsertSystemEvent(event *analytics.SystemEvent) error
}

// CreateSystemEvent creates a system-level event (no camera association)
func (s *Service) CreateSystemEvent(eventType, severity, title, message string) (*Event, error) {
	event := &Event{
		EventType: eventType,
		Severity:  severity,
		Title:     title,
		Message:   message,
		Timestamp: time.Now(),
	}

	if err := s.eventRepo.CreateEvent(event); err != nil {
		return nil, fmt.Errorf("failed to create system event: %w", err)
	}

	log.Printf("Created system event: %s (%s) - %s", eventType, severity, title)
	return event, nil
}

// GetRecentEvents retrieves recent events
func (s *Service) GetRecentEvents(limit int) ([]Event, error) {
	return s.eventRepo.GetRecentEvents(limit)
}

// GetEventsPaginated retrieves events with pagination and optional type filter
func (s *Service) GetEventsPaginated(limit, offset int, eventType string) ([]Event, int, error) {
	return s.eventRepo.GetEventsPaginated(limit, offset, eventType)
}

// Statistics methods

// GetDatabaseStats returns database statistics
func (s *Service) GetDatabaseStats() (map[string]int, error) {
	return s.db.GetDatabaseStats()
}

// GetCameraStatus returns current camera status
func (s *Service) GetCameraStatus() (*Camera, error) {
	if err := s.ensureDefaultCamera(); err != nil {
		return nil, err
	}
	return s.cameraRepo.GetCamera(s.defaultCameraID)
}

// UpdateCameraStatus updates camera status
func (s *Service) UpdateCameraStatus(status string) error {
	if err := s.ensureDefaultCamera(); err != nil {
		return err
	}
	return s.cameraRepo.UpdateCameraStatus(s.defaultCameraID, status)
}

// Cleanup methods

// CleanupOldRecordings removes old recording records and files
func (s *Service) CleanupOldRecordings(olderThanDays int) error {
	// This would implement cleanup logic
	log.Printf("TODO: Cleanup recordings older than %d days", olderThanDays)
	return nil
}

// GetStorageUsage calculates total storage usage
func (s *Service) GetStorageUsage() (int64, error) {
	var totalSize int64

	err := filepath.Walk(s.config.OutputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

// EnforceStorageLimit keeps output directory usage below maxBytes.
// Cleanup order: oldest completed recordings, then oldest standalone frames.
func (s *Service) EnforceStorageLimit(maxBytes int64) error {
	if maxBytes <= 0 {
		return nil
	}

	usage, err := s.GetStorageUsage()
	if err != nil {
		return err
	}
	if usage <= maxBytes {
		return nil
	}

	target := int64(float64(maxBytes) * 0.90)
	if target <= 0 {
		target = maxBytes
	}
	bytesToFree := usage - target
	var freed int64

	log.Printf("🧹 Storage limit exceeded: usage=%d bytes, limit=%d bytes, target=%d bytes",
		usage, maxBytes, target)

	// 1) Remove oldest completed recordings (cascade removes linked frames/events/detections).
	recordingRows, err := s.db.GetConnection().Query(`
		SELECT id, file_path, COALESCE(file_size, 0)
		FROM recordings
		WHERE status = 'completed'
		ORDER BY COALESCE(end_time, start_time, created_at) ASC
	`)
	if err != nil {
		return fmt.Errorf("failed to list recordings for cleanup: %w", err)
	}
	defer recordingRows.Close()

	for recordingRows.Next() && freed < bytesToFree {
		var id int
		var filePath string
		var fileSize int64
		if err := recordingRows.Scan(&id, &filePath, &fileSize); err != nil {
			return fmt.Errorf("failed to scan recording for cleanup: %w", err)
		}

		size := fileSize
		if st, statErr := os.Stat(filePath); statErr == nil {
			size = st.Size()
		}

		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			log.Printf("⚠️  Cleanup: could not remove recording file %s: %v", filePath, err)
			continue
		}

		if err := s.recordingRepo.DeleteRecording(id); err != nil {
			log.Printf("⚠️  Cleanup: could not delete recording row %d: %v", id, err)
			continue
		}

		freed += size
		log.Printf("🗑️  Cleanup: removed recording id=%d size=%d", id, size)
	}
	if err := recordingRows.Err(); err != nil {
		return fmt.Errorf("cleanup recording iteration failed: %w", err)
	}

	// 2) If still above budget, remove oldest standalone frames.
	if freed < bytesToFree {
		frameRows, err := s.db.GetConnection().Query(`
			SELECT id, file_path, COALESCE(file_size, 0)
			FROM frames
			WHERE recording_id IS NULL
			ORDER BY timestamp ASC
		`)
		if err != nil {
			return fmt.Errorf("failed to list standalone frames for cleanup: %w", err)
		}
		defer frameRows.Close()

		for frameRows.Next() && freed < bytesToFree {
			var id int
			var filePath string
			var fileSize int64
			if err := frameRows.Scan(&id, &filePath, &fileSize); err != nil {
				return fmt.Errorf("failed to scan frame for cleanup: %w", err)
			}

			size := fileSize
			if st, statErr := os.Stat(filePath); statErr == nil {
				size = st.Size()
			}

			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				log.Printf("⚠️  Cleanup: could not remove frame file %s: %v", filePath, err)
				continue
			}

			if _, err := s.db.GetConnection().Exec(`DELETE FROM frames WHERE id = $1`, id); err != nil {
				log.Printf("⚠️  Cleanup: could not delete frame row %d: %v", id, err)
				continue
			}

			freed += size
			log.Printf("🗑️  Cleanup: removed standalone frame id=%d size=%d", id, size)
		}
		if err := frameRows.Err(); err != nil {
			return fmt.Errorf("cleanup frame iteration failed: %w", err)
		}
	}

	newUsage, _ := s.GetStorageUsage()
	log.Printf("✅ Storage cleanup finished: freed=%d bytes, current_usage=%d bytes", freed, newUsage)
	return nil
}

// Добавить эти методы в internal/storage/service.go

// GetRecordings retrieves recordings for a camera with limit
func (s *Service) GetRecordings(cameraID int, limit int) ([]Recording, error) {
	return s.recordingRepo.GetRecordingsByCamera(cameraID, limit)
}

// GetRecording retrieves a single recording by ID
func (s *Service) GetRecording(recordingID int) (*Recording, error) {
	return s.recordingRepo.GetRecording(recordingID)
}

// GetEvent retrieves a single event by ID
func (s *Service) GetEvent(eventID int) (*Event, error) {
	return s.eventRepo.GetEvent(eventID)
}

// DeleteRecording deletes a recording from database
func (s *Service) DeleteRecording(recordingID int) error {
	return s.recordingRepo.DeleteRecording(recordingID)
}

// GetFrame retrieves a single frame by ID
func (s *Service) GetFrame(frameID int) (*Frame, error) {
	return s.frameRepo.GetFrame(frameID)
}

// GetAllFrames retrieves all frames
func (s *Service) GetAllFrames() ([]Frame, error) {
	return s.frameRepo.GetAllFrames()
}

// GetFramesPaginated retrieves frames with pagination support
func (s *Service) GetFramesPaginated(limit, offset int) ([]Frame, int, error) {
	return s.frameRepo.GetFramesPaginated(limit, offset)
}

// GetFramesByTimeRange retrieves frames within a time range
func (s *Service) GetFramesByTimeRange(cameraID int, start, end time.Time, limit int) ([]Frame, error) {
	return s.frameRepo.GetFramesByTimeRange(cameraID, start, end, limit)
}

// GetFramesWithDetection retrieves frames that have detections
func (s *Service) GetFramesWithDetection(cameraID int, limit int) ([]Frame, error) {
	query := `
		SELECT id, recording_id, camera_id, file_path, thumbnail_path, file_size,
			   width, height, timestamp, has_detection, processed, created_at
		FROM frames 
		WHERE camera_id = $1 AND has_detection = TRUE
		ORDER BY timestamp DESC 
		LIMIT $2`

	rows, err := s.db.GetConnection().Query(query, cameraID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var frames []Frame
	for rows.Next() {
		var frame Frame
		err := rows.Scan(&frame.ID, &frame.RecordingID, &frame.CameraID, &frame.FilePath,
			&frame.ThumbnailPath, &frame.FileSize, &frame.Width, &frame.Height,
			&frame.Timestamp, &frame.HasDetection, &frame.Processed, &frame.CreatedAt)
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame)
	}

	return frames, rows.Err()
}

// GetDetectionsByFrame retrieves all detections for a specific frame
func (s *Service) GetDetectionsByFrame(frameID int) ([]Detection, error) {
	query := `
		SELECT id, frame_id, object_type, confidence, bounding_box, timestamp, created_at
		FROM detections
		WHERE frame_id = $1
		ORDER BY confidence DESC`

	rows, err := s.db.GetConnection().Query(query, frameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var detections []Detection
	for rows.Next() {
		var detection Detection
		err := rows.Scan(&detection.ID, &detection.FrameID, &detection.ObjectType,
			&detection.Confidence, &detection.BoundingBox, &detection.Timestamp, &detection.CreatedAt)
		if err != nil {
			return nil, err
		}
		detections = append(detections, detection)
	}

	return detections, rows.Err()
}

// SaveDetection saves a detection result to database
func (s *Service) SaveDetection(frameID int, objectType string, confidence float64, boundingBox string) (*Detection, error) {
	detection := &Detection{
		FrameID:     frameID,
		ObjectType:  objectType,
		Confidence:  confidence,
		BoundingBox: boundingBox,
		Timestamp:   time.Now(),
	}

	query := `
		INSERT INTO detections (frame_id, object_type, confidence, bounding_box, timestamp)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := s.db.GetConnection().QueryRow(query, detection.FrameID, detection.ObjectType,
		detection.Confidence, detection.BoundingBox, detection.Timestamp).
		Scan(&detection.ID, &detection.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to save detection: %w", err)
	}

	return detection, nil
}

// GetRecentDetections retrieves recent detections across all cameras
func (s *Service) GetRecentDetections(limit int) ([]Detection, error) {
	query := `
		SELECT id, frame_id, object_type, confidence, bounding_box, timestamp, created_at
		FROM detections
		ORDER BY timestamp DESC
		LIMIT $1`

	rows, err := s.db.GetConnection().Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var detections []Detection
	for rows.Next() {
		var detection Detection
		err := rows.Scan(&detection.ID, &detection.FrameID, &detection.ObjectType,
			&detection.Confidence, &detection.BoundingBox, &detection.Timestamp, &detection.CreatedAt)
		if err != nil {
			return nil, err
		}
		detections = append(detections, detection)
	}

	return detections, rows.Err()
}

// GetDetectionStats returns statistics about detections
func (s *Service) GetDetectionStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total detections
	var totalDetections int
	err := s.db.GetConnection().QueryRow("SELECT COUNT(*) FROM detections").Scan(&totalDetections)
	if err != nil {
		return nil, err
	}
	stats["total_detections"] = totalDetections

	// Detections by object type
	objectCounts := make(map[string]int)
	rows, err := s.db.GetConnection().Query(`
		SELECT object_type, COUNT(*) as count
		FROM detections
		GROUP BY object_type
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var objectType string
		var count int
		if err := rows.Scan(&objectType, &count); err != nil {
			return nil, err
		}
		objectCounts[objectType] = count
	}
	stats["by_object_type"] = objectCounts

	// Average confidence
	var avgConfidence float64
	err = s.db.GetConnection().QueryRow("SELECT AVG(confidence) FROM detections").Scan(&avgConfidence)
	if err == nil {
		stats["avg_confidence"] = avgConfidence
	}

	// Detections today
	var detectionsToday int
	err = s.db.GetConnection().QueryRow(`
		SELECT COUNT(*) FROM detections
		WHERE DATE(timestamp) = CURRENT_DATE
	`).Scan(&detectionsToday)
	if err == nil {
		stats["detections_today"] = detectionsToday
	}

	return stats, nil
}

// UpdateCameraPing updates camera's last_ping timestamp
func (s *Service) UpdateCameraPing() error {
	query := `
		UPDATE cameras 
		SET last_ping = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err := s.db.GetConnection().Exec(query, s.defaultCameraID)
	if err != nil {
		return fmt.Errorf("failed to update camera ping: %w", err)
	}

	return nil
}
