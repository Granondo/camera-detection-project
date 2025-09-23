package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// Database wraps the sql.DB connection
type Database struct {
	conn *sql.DB
	cfg  *DatabaseConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)
	conn.SetConnMaxLifetime(cfg.MaxLifetime)

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connected to PostgreSQL database: %s@%s:%d/%s",
		cfg.User, cfg.Host, cfg.Port, cfg.Database)

	return &Database{
		conn: conn,
		cfg:  cfg,
	}, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// GetConnection returns the underlying sql.DB connection
func (db *Database) GetConnection() *sql.DB {
	return db.conn
}

// Ping checks database connectivity
func (db *Database) Ping() error {
	return db.conn.Ping()
}

// ExecuteInTransaction executes a function within a database transaction
func (db *Database) ExecuteInTransaction(fn func(*sql.Tx) error) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %w (original error: %v)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreateTables creates all necessary database tables (for development/testing)
func (db *Database) CreateTables() error {
	queries := []string{
		createCamerasTable,
		createRecordingsTable,
		createFramesTable,
		createDetectionsTable,
		createEventsTable,
		createSystemStatsTable,
		createIndexes,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	log.Println("Database tables created successfully")
	return nil
}

// Database table creation queries
const createCamerasTable = `
CREATE TABLE IF NOT EXISTS cameras (
	id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	rtsp_url TEXT NOT NULL,
	username VARCHAR(100),
	password VARCHAR(100),
	status VARCHAR(20) DEFAULT 'inactive',
	last_ping TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createRecordingsTable = `
CREATE TABLE IF NOT EXISTS recordings (
	id SERIAL PRIMARY KEY,
	camera_id INTEGER REFERENCES cameras(id) ON DELETE CASCADE,
	file_path TEXT NOT NULL,
	file_size BIGINT DEFAULT 0,
	duration INTEGER DEFAULT 0,
	start_time TIMESTAMP NOT NULL,
	end_time TIMESTAMP,
	quality VARCHAR(10),
	codec VARCHAR(20),
	status VARCHAR(20) DEFAULT 'recording',
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	archived_at TIMESTAMP
);`

const createFramesTable = `
CREATE TABLE IF NOT EXISTS frames (
	id SERIAL PRIMARY KEY,
	recording_id INTEGER REFERENCES recordings(id) ON DELETE CASCADE,
	camera_id INTEGER REFERENCES cameras(id) ON DELETE CASCADE,
	file_path TEXT NOT NULL,
	thumbnail_path TEXT,
	file_size INTEGER DEFAULT 0,
	width INTEGER DEFAULT 0,
	height INTEGER DEFAULT 0,
	timestamp TIMESTAMP NOT NULL,
	has_detection BOOLEAN DEFAULT FALSE,
	processed BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createDetectionsTable = `
CREATE TABLE IF NOT EXISTS detections (
	id SERIAL PRIMARY KEY,
	frame_id INTEGER REFERENCES frames(id) ON DELETE CASCADE,
	object_type VARCHAR(50) NOT NULL,
	confidence DECIMAL(3,2) NOT NULL,
	bounding_box TEXT,
	timestamp TIMESTAMP NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createEventsTable = `
CREATE TABLE IF NOT EXISTS events (
	id SERIAL PRIMARY KEY,
	camera_id INTEGER REFERENCES cameras(id) ON DELETE SET NULL,
	event_type VARCHAR(50) NOT NULL,
	severity VARCHAR(20) NOT NULL,
	title VARCHAR(200) NOT NULL,
	message TEXT,
	metadata TEXT,
	notified BOOLEAN DEFAULT FALSE,
	resolved BOOLEAN DEFAULT FALSE,
	timestamp TIMESTAMP NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	resolved_at TIMESTAMP
);`

const createSystemStatsTable = `
CREATE TABLE IF NOT EXISTS system_stats (
	id SERIAL PRIMARY KEY,
	date DATE NOT NULL UNIQUE,
	total_recordings INTEGER DEFAULT 0,
	total_frames INTEGER DEFAULT 0,
	total_detections INTEGER DEFAULT 0,
	storage_used_bytes BIGINT DEFAULT 0,
	uptime_seconds INTEGER DEFAULT 0,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_recordings_camera_start_time ON recordings(camera_id, start_time);
CREATE INDEX IF NOT EXISTS idx_recordings_status ON recordings(status);
CREATE INDEX IF NOT EXISTS idx_frames_camera_timestamp ON frames(camera_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_frames_has_detection ON frames(has_detection);
CREATE INDEX IF NOT EXISTS idx_detections_frame_id ON detections(frame_id);
CREATE INDEX IF NOT EXISTS idx_detections_object_type ON detections(object_type);
CREATE INDEX IF NOT EXISTS idx_events_camera_timestamp ON events(camera_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_events_severity_notified ON events(severity, notified);
CREATE INDEX IF NOT EXISTS idx_system_stats_date ON system_stats(date);
`

// GetDatabaseStats returns basic database statistics
func (db *Database) GetDatabaseStats() (map[string]int, error) {
	stats := make(map[string]int)

	queries := map[string]string{
		"cameras":     "SELECT COUNT(*) FROM cameras",
		"recordings":  "SELECT COUNT(*) FROM recordings",
		"frames":      "SELECT COUNT(*) FROM frames",
		"detections":  "SELECT COUNT(*) FROM detections",
		"events":      "SELECT COUNT(*) FROM events",
	}

	for name, query := range queries {
		var count int
		if err := db.conn.QueryRow(query).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to get %s count: %w", name, err)
		}
		stats[name] = count
	}

	return stats, nil
}