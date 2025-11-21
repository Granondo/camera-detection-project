package queue

import "time"

// FrameMessage сообщение с информацией о кадре для детекции
type FrameMessage struct {
	FrameID     int       `json:"frame_id"`
	RecordingID *int      `json:"recording_id,omitempty"`
	CameraID    int       `json:"camera_id"`
	FilePath    string    `json:"file_path"`
	Timestamp   time.Time `json:"timestamp"`
	FrameNumber int       `json:"frame_number"`
}

// EventMessage сообщение о событии
type EventMessage struct {
	EventType  string                 `json:"event_type"`
	Severity   string                 `json:"severity"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Timestamp  time.Time              `json:"timestamp"`
	CameraID   *int                   `json:"camera_id,omitempty"`
	FrameID    *int                   `json:"frame_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// DetectionResult результат детекции (ответ от worker)
type DetectionResult struct {
	FrameID          int                    `json:"frame_id"`
	Success          bool                   `json:"success"`
	TotalObjects     int                    `json:"total_objects"`
	ProcessingTimeMs float64                `json:"processing_time_ms"`
	Detections       []DetectionObject      `json:"detections"`
	Error            string                 `json:"error,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// DetectionObject обнаруженный объект
type DetectionObject struct {
	Class      string  `json:"class"`
	ClassID    int     `json:"class_id"`
	Confidence float64 `json:"confidence"`
	BBox       BBox    `json:"bbox"`
}

// BBox bounding box
type BBox struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}