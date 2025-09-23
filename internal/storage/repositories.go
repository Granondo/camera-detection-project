package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// CameraRepository handles camera data operations
type CameraRepository struct {
	db *Database
}

// RecordingRepository handles recording data operations
type RecordingRepository struct {
	db *Database
}

// FrameRepository handles frame data operations
type FrameRepository struct {
	db *Database
}

// EventRepository handles event data operations
type EventRepository struct {
	db *Database
}

// NewCameraRepository creates a new camera repository
func NewCameraRepository(db *Database) *CameraRepository {
	return &CameraRepository{db: db}
}

// NewRecordingRepository creates a new recording repository
func NewRecordingRepository(db *Database) *RecordingRepository {
	return &RecordingRepository{db: db}
}

// NewFrameRepository creates a new frame repository
func NewFrameRepository(db *Database) *FrameRepository {
	return &FrameRepository{db: db}
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *Database) *EventRepository {
	return &EventRepository{db: db}
}

// Camera Repository Methods

// CreateCamera creates a new camera record
func (r *CameraRepository) CreateCamera(camera *Camera) error {
	query := `
		INSERT INTO cameras (name, rtsp_url, username, password, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	return r.db.conn.QueryRow(query, camera.Name, camera.RTSPURL, camera.Username,
		camera.Password, camera.Status).Scan(&camera.ID, &camera.CreatedAt, &camera.UpdatedAt)
}

// GetCamera retrieves a camera by ID
func (r *CameraRepository) GetCamera(id int) (*Camera, error) {
	camera := &Camera{}
	query := `
		SELECT id, name, rtsp_url, username, password, status, last_ping, created_at, updated_at
		FROM cameras WHERE id = $1`

	err := r.db.conn.QueryRow(query, id).Scan(&camera.ID, &camera.Name, &camera.RTSPURL,
		&camera.Username, &camera.Password, &camera.Status, &camera.LastPing,
		&camera.CreatedAt, &camera.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("camera not found")
	}
	return camera, err
}

// UpdateCameraStatus updates camera status and last ping
func (r *CameraRepository) UpdateCameraStatus(id int, status string) error {
	query := `
		UPDATE cameras 
		SET status = $1, last_ping = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`

	_, err := r.db.conn.Exec(query, status, id)
	return err
}

// GetAllCameras retrieves all cameras
func (r *CameraRepository) GetAllCameras() ([]Camera, error) {
	query := `
		SELECT id, name, rtsp_url, username, password, status, last_ping, created_at, updated_at
		FROM cameras ORDER BY created_at`

	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cameras []Camera
	for rows.Next() {
		var camera Camera
		err := rows.Scan(&camera.ID, &camera.Name, &camera.RTSPURL, &camera.Username,
			&camera.Password, &camera.Status, &camera.LastPing, &camera.CreatedAt, &camera.UpdatedAt)
		if err != nil {
			return nil, err
		}
		cameras = append(cameras, camera)
	}

	return cameras, rows.Err()
}

// Recording Repository Methods

// CreateRecording creates a new recording record
func (r *RecordingRepository) CreateRecording(recording *Recording) error {
	query := `
		INSERT INTO recordings (camera_id, file_path, start_time, quality, codec, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return r.db.conn.QueryRow(query, recording.CameraID, recording.FilePath,
		recording.StartTime, recording.Quality, recording.Codec, recording.Status).
		Scan(&recording.ID, &recording.CreatedAt)
}

// UpdateRecording updates recording information
func (r *RecordingRepository) UpdateRecording(recording *Recording) error {
	query := `
		UPDATE recordings 
		SET file_size = $1, duration = $2, end_time = $3, status = $4
		WHERE id = $5`

	_, err := r.db.conn.Exec(query, recording.FileSize, recording.Duration,
		recording.EndTime, recording.Status, recording.ID)
	return err
}

// GetRecording retrieves a recording by ID
func (r *RecordingRepository) GetRecording(id int) (*Recording, error) {
	recording := &Recording{}
	query := `
		SELECT id, camera_id, file_path, file_size, duration, start_time, end_time,
			   quality, codec, status, created_at, archived_at
		FROM recordings WHERE id = $1`

	err := r.db.conn.QueryRow(query, id).Scan(&recording.ID, &recording.CameraID,
		&recording.FilePath, &recording.FileSize, &recording.Duration, &recording.StartTime,
		&recording.EndTime, &recording.Quality, &recording.Codec, &recording.Status,
		&recording.CreatedAt, &recording.ArchivedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("recording not found")
	}
	return recording, err
}

// GetRecordingsByCamera retrieves recordings for a specific camera
func (r *RecordingRepository) GetRecordingsByCamera(cameraID int, limit int) ([]Recording, error) {
	query := `
		SELECT id, camera_id, file_path, file_size, duration, start_time, end_time,
			   quality, codec, status, created_at, archived_at
		FROM recordings 
		WHERE camera_id = $1 
		ORDER BY start_time DESC 
		LIMIT $2`

	rows, err := r.db.conn.Query(query, cameraID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recordings []Recording
	for rows.Next() {
		var recording Recording
		err := rows.Scan(&recording.ID, &recording.CameraID, &recording.FilePath,
			&recording.FileSize, &recording.Duration, &recording.StartTime, &recording.EndTime,
			&recording.Quality, &recording.Codec, &recording.Status, &recording.CreatedAt,
			&recording.ArchivedAt)
		if err != nil {
			return nil, err
		}
		recordings = append(recordings, recording)
	}

	return recordings, rows.Err()
}

// Frame Repository Methods

// CreateFrame creates a new frame record
func (r *FrameRepository) CreateFrame(frame *Frame) error {
	query := `
		INSERT INTO frames (recording_id, camera_id, file_path, file_size, width, height, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	return r.db.conn.QueryRow(query, frame.RecordingID, frame.CameraID, frame.FilePath,
		frame.FileSize, frame.Width, frame.Height, frame.Timestamp).
		Scan(&frame.ID, &frame.CreatedAt)
}

// UpdateFrame updates frame information
func (r *FrameRepository) UpdateFrame(frame *Frame) error {
	query := `
		UPDATE frames 
		SET thumbnail_path = $1, has_detection = $2, processed = $3
		WHERE id = $4`

	_, err := r.db.conn.Exec(query, frame.ThumbnailPath, frame.HasDetection,
		frame.Processed, frame.ID)
	return err
}

// GetFrame retrieves a frame by ID
func (r *FrameRepository) GetFrame(id int) (*Frame, error) {
	frame := &Frame{}
	query := `
		SELECT id, recording_id, camera_id, file_path, thumbnail_path, file_size,
			   width, height, timestamp, has_detection, processed, created_at
		FROM frames WHERE id = $1`

	err := r.db.conn.QueryRow(query, id).Scan(&frame.ID, &frame.RecordingID,
		&frame.CameraID, &frame.FilePath, &frame.ThumbnailPath, &frame.FileSize,
		&frame.Width, &frame.Height, &frame.Timestamp, &frame.HasDetection,
		&frame.Processed, &frame.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("frame not found")
	}
	return frame, err
}

// GetFramesByTimeRange retrieves frames within a time range
func (r *FrameRepository) GetFramesByTimeRange(cameraID int, start, end time.Time, limit int) ([]Frame, error) {
	query := `
		SELECT id, recording_id, camera_id, file_path, thumbnail_path, file_size,
			   width, height, timestamp, has_detection, processed, created_at
		FROM frames 
		WHERE camera_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC 
		LIMIT $4`

	rows, err := r.db.conn.Query(query, cameraID, start, end, limit)
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

// GetUnprocessedFrames retrieves frames that haven't been processed yet
func (r *FrameRepository) GetUnprocessedFrames(limit int) ([]Frame, error) {
	query := `
		SELECT id, recording_id, camera_id, file_path, thumbnail_path, file_size,
			   width, height, timestamp, has_detection, processed, created_at
		FROM frames 
		WHERE processed = FALSE
		ORDER BY created_at ASC 
		LIMIT $1`

	rows, err := r.db.conn.Query(query, limit)
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

// Event Repository Methods

// CreateEvent creates a new event record
func (r *EventRepository) CreateEvent(event *Event) error {
	query := `
		INSERT INTO events (camera_id, event_type, severity, title, message, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	return r.db.conn.QueryRow(query, event.CameraID, event.EventType, event.Severity,
		event.Title, event.Message, event.Metadata, event.Timestamp).
		Scan(&event.ID, &event.CreatedAt)
}

// GetRecentEvents retrieves recent events
func (r *EventRepository) GetRecentEvents(limit int) ([]Event, error) {
	query := `
		SELECT id, camera_id, event_type, severity, title, message, metadata,
			   notified, resolved, timestamp, created_at, resolved_at
		FROM events 
		ORDER BY timestamp DESC 
		LIMIT $1`

	rows, err := r.db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(&event.ID, &event.CameraID, &event.EventType, &event.Severity,
			&event.Title, &event.Message, &event.Metadata, &event.Notified, &event.Resolved,
			&event.Timestamp, &event.CreatedAt, &event.ResolvedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

// GetUnnotifiedEvents retrieves events that haven't been notified yet
func (r *EventRepository) GetUnnotifiedEvents() ([]Event, error) {
	query := `
		SELECT id, camera_id, event_type, severity, title, message, metadata,
			   notified, resolved, timestamp, created_at, resolved_at
		FROM events 
		WHERE notified = FALSE AND (severity = 'high' OR severity = 'critical')
		ORDER BY timestamp ASC`

	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(&event.ID, &event.CameraID, &event.EventType, &event.Severity,
			&event.Title, &event.Message, &event.Metadata, &event.Notified, &event.Resolved,
			&event.Timestamp, &event.CreatedAt, &event.ResolvedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

// MarkEventNotified marks an event as notified
func (r *EventRepository) MarkEventNotified(id int) error {
	query := `UPDATE events SET notified = TRUE WHERE id = $1`
	_, err := r.db.conn.Exec(query, id)
	return err
}

// MarkEventResolved marks an event as resolved
func (r *EventRepository) MarkEventResolved(id int) error {
	query := `UPDATE events SET resolved = TRUE, resolved_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.conn.Exec(query, id)
	return err
}