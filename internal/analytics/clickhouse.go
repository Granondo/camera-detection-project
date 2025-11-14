package analytics

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// ClickHouseClient обертка для ClickHouse клиента
type ClickHouseClient struct {
	conn driver.Conn
	ctx  context.Context
}

// ClickHouseConfig конфигурация ClickHouse
type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	Debug    bool
}

// NewClickHouseClient создает новый ClickHouse клиент
func NewClickHouseClient(cfg *ClickHouseConfig) (*ClickHouseClient, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout: 10 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		// TLS отключен - Docker ClickHouse без SSL
		Debug: cfg.Debug,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	ctx := context.Background()

	// Проверка подключения
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	log.Printf("✅ Connected to ClickHouse at %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)

	return &ClickHouseClient{
		conn: conn,
		ctx:  ctx,
	}, nil
}

// Close закрывает соединение
func (c *ClickHouseClient) Close() error {
	return c.conn.Close()
}

// Ping проверяет соединение
func (c *ClickHouseClient) Ping() error {
	return c.conn.Ping(c.ctx)
}

// Detection структура для детекции
type Detection struct {
	Timestamp        time.Time
	CameraID         uint32
	RecordingID      uint64
	FrameID          uint64
	ObjectClass      string
	Confidence       float32
	BBoxX1           float32
	BBoxY1           float32
	BBoxX2           float32
	BBoxY2           float32
	ModelVersion     string
	ProcessingTimeMs uint32
	TrackingID       *string
}

// InsertDetection вставляет детекцию
func (c *ClickHouseClient) InsertDetection(d *Detection) error {
	query := `
		INSERT INTO detections (
			timestamp, camera_id, recording_id, frame_id,
			object_class, confidence,
			bbox_x1, bbox_y1, bbox_x2, bbox_y2,
			model_version, processing_time_ms, tracking_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return c.conn.Exec(c.ctx, query,
		d.Timestamp, d.CameraID, d.RecordingID, d.FrameID,
		d.ObjectClass, d.Confidence,
		d.BBoxX1, d.BBoxY1, d.BBoxX2, d.BBoxY2,
		d.ModelVersion, d.ProcessingTimeMs, d.TrackingID,
	)
}

// InsertDetectionsBatch вставляет несколько детекций за раз
func (c *ClickHouseClient) InsertDetectionsBatch(detections []Detection) error {
	if len(detections) == 0 {
		return nil
	}

	batch, err := c.conn.PrepareBatch(c.ctx, `
		INSERT INTO detections (
			timestamp, camera_id, recording_id, frame_id,
			object_class, confidence,
			bbox_x1, bbox_y1, bbox_x2, bbox_y2,
			model_version, processing_time_ms, tracking_id
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, d := range detections {
		err := batch.Append(
			d.Timestamp, d.CameraID, d.RecordingID, d.FrameID,
			d.ObjectClass, d.Confidence,
			d.BBoxX1, d.BBoxY1, d.BBoxX2, d.BBoxY2,
			d.ModelVersion, d.ProcessingTimeMs, d.TrackingID,
		)
		if err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	log.Printf("✅ Inserted %d detections to ClickHouse", len(detections))
	return nil
}

// SystemEvent структура для события системы
type SystemEvent struct {
	Timestamp   time.Time
	EventType   string
	Severity    string
	Service     string
	Title       string
	Message     string
	StackTrace  *string
	CameraID    *uint32
	RecordingID *uint64
	FrameID     *uint64
	Metadata    string
}

// InsertSystemEvent вставляет событие системы
func (c *ClickHouseClient) InsertSystemEvent(e *SystemEvent) error {
	query := `
		INSERT INTO system_events (
			timestamp, event_type, severity, service,
			title, message, stack_trace,
			camera_id, recording_id, frame_id, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return c.conn.Exec(c.ctx, query,
		e.Timestamp, e.EventType, e.Severity, e.Service,
		e.Title, e.Message, e.StackTrace,
		e.CameraID, e.RecordingID, e.FrameID, e.Metadata,
	)
}

// SystemMetrics структура для метрик системы
type SystemMetrics struct {
	Timestamp             time.Time
	DetectionLatencyMs    float32
	FramesProcessedPerSec float32
	DetectionsPerFrame    float32
	CPUPercent            float32
	MemoryPercent         float32
	MemoryUsedMb          float32
	DiskUsagePercent      float32
	CameraID              uint32
	CameraFPS             float32
	CameraBitrateKbps     uint32
	APIRequestsPerSec     float32
	APIAvgResponseMs      float32
	RedisHitRate          float32
	RedisMemoryMb         float32
}

// InsertSystemMetrics вставляет метрики системы
func (c *ClickHouseClient) InsertSystemMetrics(m *SystemMetrics) error {
	query := `
		INSERT INTO system_metrics (
			timestamp,
			detection_latency_ms, frames_processed_per_sec, detections_per_frame,
			cpu_percent, memory_percent, memory_used_mb, disk_usage_percent,
			camera_id, camera_fps, camera_bitrate_kbps,
			api_requests_per_sec, api_avg_response_ms,
			redis_hit_rate, redis_memory_mb
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return c.conn.Exec(c.ctx, query,
		m.Timestamp,
		m.DetectionLatencyMs, m.FramesProcessedPerSec, m.DetectionsPerFrame,
		m.CPUPercent, m.MemoryPercent, m.MemoryUsedMb, m.DiskUsagePercent,
		m.CameraID, m.CameraFPS, m.CameraBitrateKbps,
		m.APIRequestsPerSec, m.APIAvgResponseMs,
		m.RedisHitRate, m.RedisMemoryMb,
	)
}

// DetectionStats статистика по детекциям
type DetectionStats struct {
	Hour            time.Time
	ObjectClass     string
	TotalDetections uint64
	AvgConfidence   float64
	MinConfidence   float64
	MaxConfidence   float64
}

// GetDetectionsByHour получает детекции по часам
func (c *ClickHouseClient) GetDetectionsByHour(cameraID uint32, hours int) ([]DetectionStats, error) {
	query := `
		SELECT 
			toStartOfHour(timestamp) as hour,
			object_class,
			count() as total_detections,
			avg(confidence) as avg_confidence,
			min(confidence) as min_confidence,
			max(confidence) as max_confidence
		FROM detections
		WHERE camera_id = ?
		  AND timestamp >= now() - INTERVAL ? HOUR
		GROUP BY hour, object_class
		ORDER BY hour DESC, total_detections DESC
	`

	rows, err := c.conn.Query(c.ctx, query, cameraID, hours)
	if err != nil {
		return nil, fmt.Errorf("failed to query detections by hour: %w", err)
	}
	defer rows.Close()

	var results []DetectionStats
	for rows.Next() {
		var stat DetectionStats
		if err := rows.Scan(
			&stat.Hour,
			&stat.ObjectClass,
			&stat.TotalDetections,
			&stat.AvgConfidence,
			&stat.MinConfidence,
			&stat.MaxConfidence,
		); err != nil {
			return nil, err
		}
		results = append(results, stat)
	}

	return results, rows.Err()
}

// TopDetectedObjects топ обнаруженных объектов
type TopDetectedObjects struct {
	ObjectClass     string
	TotalDetections uint64
	AvgConfidence   float64
}

// GetTopDetectedObjects получает топ обнаруженных объектов
func (c *ClickHouseClient) GetTopDetectedObjects(cameraID uint32, days int, limit int) ([]TopDetectedObjects, error) {
	query := `
		SELECT 
			object_class,
			count() as total_detections,
			avg(confidence) as avg_confidence
		FROM detections
		WHERE camera_id = ?
		  AND date >= today() - INTERVAL ? DAY
		GROUP BY object_class
		ORDER BY total_detections DESC
		LIMIT ?
	`

	rows, err := c.conn.Query(c.ctx, query, cameraID, days, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top objects: %w", err)
	}
	defer rows.Close()

	var results []TopDetectedObjects
	for rows.Next() {
		var obj TopDetectedObjects
		if err := rows.Scan(
			&obj.ObjectClass,
			&obj.TotalDetections,
			&obj.AvgConfidence,
		); err != nil {
			return nil, err
		}
		results = append(results, obj)
	}

	return results, rows.Err()
}

// GetTotalDetectionsCount получает общее количество детекций
func (c *ClickHouseClient) GetTotalDetectionsCount(cameraID uint32) (uint64, error) {
	var count uint64
	query := `SELECT count() FROM detections WHERE camera_id = ?`

	if err := c.conn.QueryRow(c.ctx, query, cameraID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get total detections count: %w", err)
	}

	return count, nil
}

// GetDetectionsCountToday получает количество детекций за сегодня
func (c *ClickHouseClient) GetDetectionsCountToday(cameraID uint32) (uint64, error) {
	var count uint64
	query := `
		SELECT count() 
		FROM detections 
		WHERE camera_id = ? AND date = today()
	`

	if err := c.conn.QueryRow(c.ctx, query, cameraID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get today's detections count: %w", err)
	}

	return count, nil
}
