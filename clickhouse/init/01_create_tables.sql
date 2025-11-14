-- Создаем базу данных если её нет
CREATE DATABASE IF NOT EXISTS surveillance;

-- Используем базу
USE surveillance;

-- Таблица для детальных детекций
CREATE TABLE IF NOT EXISTS detections (
    -- Временные метки
    timestamp DateTime64(3) CODEC(Delta, ZSTD),
    date Date DEFAULT toDate(timestamp),
    
    -- ID связей
    camera_id UInt32 CODEC(ZSTD),
    recording_id UInt64 CODEC(ZSTD),
    frame_id UInt64 CODEC(ZSTD),
    
    -- Информация о детекции
    object_class LowCardinality(String) CODEC(ZSTD),
    confidence Float32 CODEC(ZSTD),
    
    -- Bounding box
    bbox_x1 Float32 CODEC(ZSTD),
    bbox_y1 Float32 CODEC(ZSTD),
    bbox_x2 Float32 CODEC(ZSTD),
    bbox_y2 Float32 CODEC(ZSTD),
    
    -- Дополнительная информация
    model_version LowCardinality(String) DEFAULT 'yolov8n' CODEC(ZSTD),
    processing_time_ms UInt32 CODEC(ZSTD),
    
    -- Tracking (опционально)
    tracking_id Nullable(String) CODEC(ZSTD),
    
    -- Индексы
    INDEX idx_object_class object_class TYPE minmax GRANULARITY 4,
    INDEX idx_confidence confidence TYPE minmax GRANULARITY 4
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (camera_id, date, timestamp)
TTL date + INTERVAL 90 DAY  -- Автоматическое удаление через 90 дней
SETTINGS index_granularity = 8192;

-- Таблица для событий системы
CREATE TABLE IF NOT EXISTS system_events (
    -- Временные метки
    timestamp DateTime64(3) CODEC(Delta, ZSTD),
    date Date DEFAULT toDate(timestamp),
    
    -- Тип и серьезность
    event_type LowCardinality(String) CODEC(ZSTD),
    severity LowCardinality(String) CODEC(ZSTD),
    service LowCardinality(String) DEFAULT 'camera-service' CODEC(ZSTD),
    
    -- Сообщение и детали
    title String CODEC(ZSTD),
    message String CODEC(ZSTD),
    stack_trace Nullable(String) CODEC(ZSTD),
    
    -- Контекст
    camera_id Nullable(UInt32) CODEC(ZSTD),
    recording_id Nullable(UInt64) CODEC(ZSTD),
    frame_id Nullable(UInt64) CODEC(ZSTD),
    
    -- Метаданные (JSON)
    metadata String DEFAULT '{}' CODEC(ZSTD),
    
    -- Индексы
    INDEX idx_severity severity TYPE set(4) GRANULARITY 4,
    INDEX idx_event_type event_type TYPE set(10) GRANULARITY 4
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, timestamp)
TTL date + INTERVAL 180 DAY  -- Храним 6 месяцев
SETTINGS index_granularity = 8192;

-- Таблица для метрик производительности
CREATE TABLE IF NOT EXISTS system_metrics (
    -- Временные метки
    timestamp DateTime CODEC(Delta, ZSTD),
    date Date DEFAULT toDate(timestamp),
    
    -- Метрики детекции
    detection_latency_ms Float32 CODEC(ZSTD),
    frames_processed_per_sec Float32 CODEC(ZSTD),
    detections_per_frame Float32 CODEC(ZSTD),
    
    -- Метрики системы
    cpu_percent Float32 CODEC(ZSTD),
    memory_percent Float32 CODEC(ZSTD),
    memory_used_mb Float32 CODEC(ZSTD),
    disk_usage_percent Float32 CODEC(ZSTD),
    
    -- Метрики камеры
    camera_id UInt32 CODEC(ZSTD),
    camera_fps Float32 CODEC(ZSTD),
    camera_bitrate_kbps UInt32 CODEC(ZSTD),
    
    -- Метрики API
    api_requests_per_sec Float32 DEFAULT 0 CODEC(ZSTD),
    api_avg_response_ms Float32 DEFAULT 0 CODEC(ZSTD),
    
    -- Redis метрики
    redis_hit_rate Float32 DEFAULT 0 CODEC(ZSTD),
    redis_memory_mb Float32 DEFAULT 0 CODEC(ZSTD)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (camera_id, timestamp)
TTL date + INTERVAL 30 DAY  -- Храним месяц
SETTINGS index_granularity = 8192;

-- Материализованные представления для быстрых агрегаций

-- Детекции по часам
CREATE MATERIALIZED VIEW IF NOT EXISTS detections_hourly
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
GROUP BY hour, camera_id, object_class;

-- Детекции по дням
CREATE MATERIALIZED VIEW IF NOT EXISTS detections_daily
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
GROUP BY date, camera_id, object_class;

-- События по часам
CREATE MATERIALIZED VIEW IF NOT EXISTS events_hourly
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (event_type, severity, hour)
AS SELECT
    toStartOfHour(timestamp) as hour,
    event_type,
    severity,
    count() as total_events
FROM system_events
GROUP BY hour, event_type, severity;