package analytics

import (
	"fmt"
	"log"
)

// Migrate создает все таблицы если их нет
func (c *ClickHouseClient) Migrate() error {
	log.Println("🔄 Running ClickHouse migrations...")

	migrations := []string{
		// 1. Таблица детекций
		`CREATE TABLE IF NOT EXISTS detections (
			timestamp DateTime64(3) CODEC(Delta, ZSTD),
			date Date DEFAULT toDate(timestamp),
			camera_id UInt32 CODEC(ZSTD),
			recording_id UInt64 CODEC(ZSTD),
			frame_id UInt64 CODEC(ZSTD),
			object_class LowCardinality(String) CODEC(ZSTD),
			confidence Float32 CODEC(ZSTD),
			bbox_x1 Float32 CODEC(ZSTD),
			bbox_y1 Float32 CODEC(ZSTD),
			bbox_x2 Float32 CODEC(ZSTD),
			bbox_y2 Float32 CODEC(ZSTD),
			model_version LowCardinality(String) DEFAULT 'yolov8n' CODEC(ZSTD),
			processing_time_ms UInt32 CODEC(ZSTD),
			tracking_id Nullable(String) CODEC(ZSTD),
			INDEX idx_object_class object_class TYPE minmax GRANULARITY 4,
			INDEX idx_confidence confidence TYPE minmax GRANULARITY 4
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(date)
		ORDER BY (camera_id, date, timestamp)
		TTL date + INTERVAL 90 DAY
		SETTINGS index_granularity = 8192`,

		// 2. Таблица событий системы
		`CREATE TABLE IF NOT EXISTS system_events (
			timestamp DateTime64(3) CODEC(Delta, ZSTD),
			date Date DEFAULT toDate(timestamp),
			event_type LowCardinality(String) CODEC(ZSTD),
			severity LowCardinality(String) CODEC(ZSTD),
			service LowCardinality(String) DEFAULT 'camera-service' CODEC(ZSTD),
			title String CODEC(ZSTD),
			message String CODEC(ZSTD),
			stack_trace Nullable(String) CODEC(ZSTD),
			camera_id Nullable(UInt32) CODEC(ZSTD),
			recording_id Nullable(UInt64) CODEC(ZSTD),
			frame_id Nullable(UInt64) CODEC(ZSTD),
			metadata String DEFAULT '{}' CODEC(ZSTD),
			INDEX idx_severity severity TYPE set(4) GRANULARITY 4,
			INDEX idx_event_type event_type TYPE set(10) GRANULARITY 4
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(date)
		ORDER BY (date, timestamp)
		TTL date + INTERVAL 180 DAY
		SETTINGS index_granularity = 8192`,

		// 3. Таблица метрик
		`CREATE TABLE IF NOT EXISTS system_metrics (
			timestamp DateTime CODEC(Delta, ZSTD),
			date Date DEFAULT toDate(timestamp),
			detection_latency_ms Float32 CODEC(ZSTD),
			frames_processed_per_sec Float32 CODEC(ZSTD),
			detections_per_frame Float32 CODEC(ZSTD),
			cpu_percent Float32 CODEC(ZSTD),
			memory_percent Float32 CODEC(ZSTD),
			memory_used_mb Float32 CODEC(ZSTD),
			disk_usage_percent Float32 CODEC(ZSTD),
			camera_id UInt32 CODEC(ZSTD),
			camera_fps Float32 CODEC(ZSTD),
			camera_bitrate_kbps UInt32 CODEC(ZSTD),
			api_requests_per_sec Float32 DEFAULT 0 CODEC(ZSTD),
			api_avg_response_ms Float32 DEFAULT 0 CODEC(ZSTD),
			redis_hit_rate Float32 DEFAULT 0 CODEC(ZSTD),
			redis_memory_mb Float32 DEFAULT 0 CODEC(ZSTD)
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(date)
		ORDER BY (camera_id, timestamp)
		TTL date + INTERVAL 30 DAY
		SETTINGS index_granularity = 8192`,

		// 4. Материализованное представление - детекции по часам
		`CREATE MATERIALIZED VIEW IF NOT EXISTS detections_hourly
		ENGINE = SummingMergeTree()
		PARTITION BY toYYYYMM(hour)
		ORDER BY (camera_id, object_class, hour)
		AS SELECT
			toStartOfHour(timestamp) as hour,
			camera_id,
			object_class,
			count() as total_detections,
			avg(confidence) as avg_confidence,
			min(confidence) as min_confidence,
			max(confidence) as max_confidence
		FROM detections
		GROUP BY hour, camera_id, object_class`,

		// 5. Материализованное представление - детекции по дням
		`CREATE MATERIALIZED VIEW IF NOT EXISTS detections_daily
		ENGINE = SummingMergeTree()
		PARTITION BY toYYYYMM(date)
		ORDER BY (camera_id, object_class, date)
		AS SELECT
			date,
			camera_id,
			object_class,
			count() as total_detections,
			avg(confidence) as avg_confidence,
			uniq(tracking_id) as unique_objects
		FROM detections
		GROUP BY date, camera_id, object_class`,

		// 6. Материализованное представление - события по часам
		`CREATE MATERIALIZED VIEW IF NOT EXISTS events_hourly
		ENGINE = SummingMergeTree()
		PARTITION BY toYYYYMM(hour)
		ORDER BY (event_type, severity, hour)
		AS SELECT
			toStartOfHour(timestamp) as hour,
			event_type,
			severity,
			count() as total_events
		FROM system_events
		GROUP BY hour, event_type, severity`,
	}

	// Выполнить все миграции
	for i, migration := range migrations {
		log.Printf("  Running migration %d/%d...", i+1, len(migrations))
		if err := c.conn.Exec(c.ctx, migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	log.Println("✅ ClickHouse migrations completed successfully")
	return nil
}

// DropAllTables удаляет все таблицы (для разработки)
func (c *ClickHouseClient) DropAllTables() error {
	log.Println("⚠️  Dropping all ClickHouse tables...")

	tables := []string{
		"events_hourly",
		"detections_daily",
		"detections_hourly",
		"system_metrics",
		"system_events",
		"detections",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		if err := c.conn.Exec(c.ctx, query); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
		log.Printf("  Dropped table: %s", table)
	}

	log.Println("✅ All tables dropped")
	return nil
}

// CheckTablesExist проверяет существуют ли таблицы
func (c *ClickHouseClient) CheckTablesExist() (bool, error) {
	var count uint64
	query := `
		SELECT count() 
		FROM system.tables 
		WHERE database = ? AND name IN ('detections', 'system_events', 'system_metrics')
	`
	
	if err := c.conn.QueryRow(c.ctx, query, "surveillance").Scan(&count); err != nil {
		return false, err
	}
	
	return count >= 3, nil
}