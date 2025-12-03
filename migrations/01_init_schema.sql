-- Database initialization script for surveillance system
-- This script creates all necessary tables for the camera detection project

-- Cameras table
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
);

-- Recordings table
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
);

-- Frames table
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
);

-- Detections table
CREATE TABLE IF NOT EXISTS detections (
	id SERIAL PRIMARY KEY,
	frame_id INTEGER REFERENCES frames(id) ON DELETE CASCADE,
	object_type VARCHAR(50) NOT NULL,
	confidence DECIMAL(3,2) NOT NULL,
	bounding_box TEXT,
	timestamp TIMESTAMP NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Events table
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
);

-- System stats table
CREATE TABLE IF NOT EXISTS system_stats (
	id SERIAL PRIMARY KEY,
	date DATE NOT NULL UNIQUE,
	total_recordings INTEGER DEFAULT 0,
	total_frames INTEGER DEFAULT 0,
	total_detections INTEGER DEFAULT 0,
	storage_used_bytes BIGINT DEFAULT 0,
	uptime_seconds INTEGER DEFAULT 0,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_recordings_camera_start_time ON recordings(camera_id, start_time);
CREATE INDEX IF NOT EXISTS idx_recordings_status ON recordings(status);
CREATE INDEX IF NOT EXISTS idx_frames_camera_timestamp ON frames(camera_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_frames_has_detection ON frames(has_detection);
CREATE INDEX IF NOT EXISTS idx_detections_frame_id ON detections(frame_id);
CREATE INDEX IF NOT EXISTS idx_detections_object_type ON detections(object_type);
CREATE INDEX IF NOT EXISTS idx_events_camera_timestamp ON events(camera_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_events_severity_notified ON events(severity, notified);
CREATE INDEX IF NOT EXISTS idx_system_stats_date ON system_stats(date);
