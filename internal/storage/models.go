package storage

import (
	"time"
)

// Camera represents a surveillance camera
type Camera struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	RTSPURL   string    `db:"rtsp_url" json:"rtsp_url"`
	Username  string    `db:"username" json:"username"`
	Password  string    `db:"password" json:"-"` // Don't expose password in JSON
	Status    string    `db:"status" json:"status"`
	LastPing  *time.Time `db:"last_ping" json:"last_ping"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Recording represents a video recording
type Recording struct {
	ID          int        `db:"id" json:"id"`
	CameraID    int        `db:"camera_id" json:"camera_id"`
	FilePath    string     `db:"file_path" json:"file_path"`
	FileSize    int64      `db:"file_size" json:"file_size"`
	Duration    int        `db:"duration" json:"duration"` // seconds
	StartTime   time.Time  `db:"start_time" json:"start_time"`
	EndTime     *time.Time `db:"end_time" json:"end_time"`
	Quality     string     `db:"quality" json:"quality"`     // 1080p, 720p, etc
	Codec       string     `db:"codec" json:"codec"`         // h264, h265, etc
	Status      string     `db:"status" json:"status"`       // recording, completed, failed
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	ArchivedAt  *time.Time `db:"archived_at" json:"archived_at"`
}

// Frame represents an extracted frame from video
type Frame struct {
	ID            int        `db:"id" json:"id"`
	RecordingID   *int       `db:"recording_id" json:"recording_id"` // Can be null for standalone frames
	CameraID      int        `db:"camera_id" json:"camera_id"`
	FilePath      string     `db:"file_path" json:"file_path"`
	ThumbnailPath *string    `db:"thumbnail_path" json:"thumbnail_path"`
	FileSize      int        `db:"file_size" json:"file_size"`
	Width         int        `db:"width" json:"width"`
	Height        int        `db:"height" json:"height"`
	Timestamp     time.Time  `db:"timestamp" json:"timestamp"`
	HasDetection  bool       `db:"has_detection" json:"has_detection"`
	Processed     bool       `db:"processed" json:"processed"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
}

// Detection represents object detection result
type Detection struct {
	ID          int     `db:"id" json:"id"`
	FrameID     int     `db:"frame_id" json:"frame_id"`
	ObjectType  string  `db:"object_type" json:"object_type"`   // person, car, cat, etc
	Confidence  float64 `db:"confidence" json:"confidence"`     // 0.0-1.0
	BoundingBox string  `db:"bounding_box" json:"bounding_box"` // JSON: {x,y,w,h}
	Timestamp   time.Time `db:"timestamp" json:"timestamp"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// Event represents system events and alerts
type Event struct {
	ID         int       `db:"id" json:"id"`
	CameraID   *int      `db:"camera_id" json:"camera_id"`
	EventType  string    `db:"event_type" json:"event_type"` // motion, person_detected, system_error, etc
	Severity   string    `db:"severity" json:"severity"`     // low, medium, high, critical
	Title      string    `db:"title" json:"title"`
	Message    string    `db:"message" json:"message"`
	Metadata   *string   `db:"metadata" json:"metadata"`     // JSON with additional data
	Notified   bool      `db:"notified" json:"notified"`
	Resolved   bool      `db:"resolved" json:"resolved"`
	Timestamp  time.Time `db:"timestamp" json:"timestamp"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	ResolvedAt *time.Time `db:"resolved_at" json:"resolved_at"`
}

// SystemStats represents system statistics
type SystemStats struct {
	ID               int       `db:"id" json:"id"`
	Date             time.Time `db:"date" json:"date"` // Daily stats
	TotalRecordings  int       `db:"total_recordings" json:"total_recordings"`
	TotalFrames      int       `db:"total_frames" json:"total_frames"`
	TotalDetections  int       `db:"total_detections" json:"total_detections"`
	StorageUsedBytes int64     `db:"storage_used_bytes" json:"storage_used_bytes"`
	UptimeSeconds    int       `db:"uptime_seconds" json:"uptime_seconds"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

// Constants for enum values
const (
	// Camera status
	CameraStatusActive   = "active"
	CameraStatusInactive = "inactive"
	CameraStatusError    = "error"

	// Recording status
	RecordingStatusRecording = "recording"
	RecordingStatusCompleted = "completed"
	RecordingStatusFailed    = "failed"

	// Event types
	EventTypeMotion         = "motion"
	EventTypePersonDetected = "person_detected"
	EventTypeSystemStart    = "system_start"
	EventTypeSystemStop     = "system_stop"
	EventTypeSystemError    = "system_error"
	EventTypeCameraOffline  = "camera_offline"
	EventTypeCameraOnline   = "camera_online"
	EventTypeStorageFull    = "storage_full"

	// Event severity
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"

	// Object types for detection
	ObjectTypePerson  = "person"
	ObjectTypeCar     = "car"
	ObjectTypeCat     = "cat"
	ObjectTypeDog     = "dog"
	ObjectTypeBird    = "bird"
	ObjectTypePackage = "package"
)

// Helper methods for models

// IsOnline checks if camera is currently online
func (c *Camera) IsOnline() bool {
	if c.LastPing == nil {
		return false
	}
	return time.Since(*c.LastPing) < 5*time.Minute
}

// IsCompleted checks if recording is completed
func (r *Recording) IsCompleted() bool {
	return r.Status == RecordingStatusCompleted && r.EndTime != nil
}

// GetDuration calculates actual duration of recording
func (r *Recording) GetDuration() time.Duration {
	if r.EndTime == nil {
		return time.Since(r.StartTime)
	}
	return r.EndTime.Sub(r.StartTime)
}

// IsHighConfidence checks if detection has high confidence
func (d *Detection) IsHighConfidence() bool {
	return d.Confidence >= 0.8
}

// IsCritical checks if event is critical
func (e *Event) IsCritical() bool {
	return e.Severity == SeverityCritical
}

// NeedsNotification checks if event needs notification
func (e *Event) NeedsNotification() bool {
	return !e.Notified && (e.Severity == SeverityHigh || e.Severity == SeverityCritical)
}