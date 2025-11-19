package analytics

import (
	"log"
	"runtime"
	"syscall"
	"time"
)

// MetricsCollector собирает системные метрики
type MetricsCollector struct {
	client   *ClickHouseClient
	cameraID uint32
	enabled  bool
}

// NewMetricsCollector создает новый коллектор метрик
func NewMetricsCollector(client *ClickHouseClient, cameraID uint32) *MetricsCollector {
	return &MetricsCollector{
		client:   client,
		cameraID: cameraID,
		enabled:  client != nil,
	}
}

// Start запускает периодический сбор метрик
func (m *MetricsCollector) Start(interval time.Duration) {
	if !m.enabled {
		log.Println("📊 Metrics collection disabled (ClickHouse not available)")
		return
	}

	log.Printf("📊 Starting metrics collection (interval: %v)", interval)
	
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			m.collectAndSend()
		}
	}()
}

// collectAndSend собирает и отправляет метрики
func (m *MetricsCollector) collectAndSend() {
	metrics := m.collectMetrics()
	
	if err := m.client.InsertSystemMetrics(metrics); err != nil {
		log.Printf("⚠️  Failed to send metrics to ClickHouse: %v", err)
	} else {
		log.Printf("📊 Metrics sent: CPU=%.1f%%, Memory=%.1f%%", 
			metrics.CPUPercent, metrics.MemoryPercent)
	}
}

// collectMetrics собирает текущие метрики системы
func (m *MetricsCollector) collectMetrics() *SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Память
	memoryUsedMb := float32(memStats.Alloc) / 1024 / 1024
	memoryTotalMb := float32(memStats.Sys) / 1024 / 1024
	memoryPercent := (memoryUsedMb / memoryTotalMb) * 100

	// CPU (примерная оценка через goroutines)
	cpuPercent := float32(runtime.NumGoroutine()) / 100.0

	// Диск
	diskUsagePercent := m.getDiskUsage()

	return &SystemMetrics{
		Timestamp:             time.Now(),
		DetectionLatencyMs:    0, // TODO: считать среднее из детекций
		FramesProcessedPerSec: 0, // TODO: считать
		DetectionsPerFrame:    0, // TODO: считать
		CPUPercent:            cpuPercent,
		MemoryPercent:         memoryPercent,
		MemoryUsedMb:          memoryUsedMb,
		DiskUsagePercent:      diskUsagePercent,
		CameraID:              m.cameraID,
		CameraFPS:             0,
		CameraBitrateKbps:     0,
		APIRequestsPerSec:     0,
		APIAvgResponseMs:      0,
		RedisHitRate:          0,
		RedisMemoryMb:         0,
	}
}

// getDiskUsage получает процент использования диска
func (m *MetricsCollector) getDiskUsage() float32 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/app/output", &stat); err != nil {
		return 0
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free

	return float32(used) / float32(total) * 100
}