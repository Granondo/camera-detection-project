package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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
func (s *Service) CreateEvent(eventType, severity, title, message string, metadata *string) error {
	event := &Event{
		CameraID:  &s.defaultCameraID,
		EventType: eventType,
		Severity:  severity,
		Title:     title,
		Message:   message,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	if err := s.eventRepo.CreateEvent(event); err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	log.Printf("Created event: %s (%s) - %s", eventType, severity, title)
	return nil
}

// CreateSystemEvent creates a system-level event (no camera association)
func (s *Service) CreateSystemEvent(eventType, severity, title, message string) error {
	event := &Event{
		EventType: eventType,
		Severity:  severity,
		Title:     title,
		Message:   message,
		Timestamp: time.Now(),
	}

	if err := s.eventRepo.CreateEvent(event); err != nil {
		return fmt.Errorf("failed to create system event: %w", err)
	}

	log.Printf("Created system event: %s (%s) - %s", eventType, severity, title)
	return nil
}

// GetRecentEvents retrieves recent events
func (s *Service) GetRecentEvents(limit int) ([]Event, error) {
	return s.eventRepo.GetRecentEvents(limit)
}

// Statistics methods

// GetDatabaseStats returns database statistics
func (s *Service) GetDatabaseStats() (map[string]int, error) {
	return s.db.GetDatabaseStats()
}

// GetCameraStatus returns current camera status
func (s *Service) GetCameraStatus() (*Camera, error) {
	return s.cameraRepo.GetCamera(s.defaultCameraID)
}

// UpdateCameraStatus updates camera status
func (s *Service) UpdateCameraStatus(status string) error {
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