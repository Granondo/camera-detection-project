package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"camera-detection-project/internal/analytics"
	"camera-detection-project/internal/cache"
	"camera-detection-project/internal/search"
	"camera-detection-project/internal/storage"
)

type Server struct {
	storage         *storage.Service
	cache           *cache.CacheService
	analyticsClient *analytics.ClickHouseClient
	searchClient    *search.ElasticsearchClient
	outputDir       string
}

// NewServer creates a new API server without cache
func NewServer(storageService *storage.Service, outputDir string) *Server {
	return &Server{
		storage:   storageService,
		cache:     nil,
		outputDir: outputDir,
	}
}

// NewServerWithCache creates a new API server with cache
func NewServerWithCache(storageService *storage.Service, cacheService *cache.CacheService, outputDir string) *Server {
	return &Server{
		storage:   storageService,
		cache:     cacheService,
		outputDir: outputDir,
	}
}

// NewServerWithAnalytics creates a new API server with analytics (ClickHouse)
func NewServerWithAnalytics(
	storageService *storage.Service,
	cacheService *cache.CacheService,
	analyticsClient *analytics.ClickHouseClient,
	searchClient *search.ElasticsearchClient,
	outputDir string,
) *Server {
	return &Server{
		storage:         storageService,
		cache:           cacheService,
		analyticsClient: analyticsClient,
		searchClient:    searchClient,
		outputDir:       outputDir,
	}
}

// Response structures
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Cached  bool        `json:"cached,omitempty"`
}

type SystemStats struct {
	Cameras           int   `json:"cameras"`
	Recordings        int   `json:"recordings"`
	Frames            int   `json:"frames"`
	Detections        int   `json:"detections"`
	Events            int   `json:"events"`
	StorageUsageBytes int64 `json:"storage_usage_bytes"`
}

type StatusResponse struct {
	SystemStatus string          `json:"system_status"`
	Camera       *storage.Camera `json:"camera"`
	Stats        SystemStats     `json:"stats"`
	Uptime       string          `json:"uptime"`
	LastUpdated  time.Time       `json:"last_updated"`
}

type RecordingResponse struct {
	Recording   *storage.Recording `json:"recording"`
	FileExists  bool               `json:"file_exists"`
	DownloadURL string             `json:"download_url"`
	StreamURL   string             `json:"stream_url,omitempty"`
}

type FrameResponse struct {
	Frame        *storage.Frame `json:"frame"`
	FileExists   bool           `json:"file_exists"`
	ImageURL     string         `json:"image_url"`
	ThumbnailURL string         `json:"thumbnail_url,omitempty"`
	HasDetection bool           `json:"has_detection"`
}

// ✅ НОВЫЕ СТРУКТУРЫ ДЛЯ АНАЛИТИКИ
type DetectionsHourlyResponse struct {
	Hour            string  `json:"hour"`
	ObjectClass     string  `json:"object_class"`
	TotalDetections uint64  `json:"total_detections"`
	AvgConfidence   float64 `json:"avg_confidence"`
	MinConfidence   float64 `json:"min_confidence"`
	MaxConfidence   float64 `json:"max_confidence"`
}

type TopObjectsResponse struct {
	ObjectClass     string  `json:"object_class"`
	TotalDetections uint64  `json:"total_detections"`
	AvgConfidence   float64 `json:"avg_confidence"`
}

type AnalyticsSummaryResponse struct {
	TotalDetections     uint64                     `json:"total_detections"`
	DetectionsToday     uint64                     `json:"detections_today"`
	TopObjects          []TopObjectsResponse       `json:"top_objects"`
	DetectionsByHour    []DetectionsHourlyResponse `json:"detections_by_hour"`
	ClickHouseAvailable bool                       `json:"clickhouse_available"`
}

// SetupRoutes configures all API routes
func (s *Server) SetupRoutes(mux *http.ServeMux) {
	// Status and Info
	mux.HandleFunc("/api/status", s.corsMiddleware(s.handleStatus))
	mux.HandleFunc("/api/health", s.corsMiddleware(s.handleHealth))

	// Recordings
	mux.HandleFunc("/api/recordings", s.corsMiddleware(s.handleRecordings))
	mux.HandleFunc("/api/recordings/", s.corsMiddleware(s.handleRecordingByID))

	// Video streaming/download
	mux.HandleFunc("/api/video/", s.corsMiddleware(s.handleVideoStream))
	mux.HandleFunc("/api/download/", s.corsMiddleware(s.handleVideoDownload))

	// Frames
	mux.HandleFunc("/api/frames", s.corsMiddleware(s.handleFrames))
	mux.HandleFunc("/api/frames/", s.corsMiddleware(s.handleFrameByID))
	mux.HandleFunc("/api/image/", s.corsMiddleware(s.handleImageServe))

	// Events
	mux.HandleFunc("/api/events", s.corsMiddleware(s.handleEvents))

	// Search (Elasticsearch)
	mux.HandleFunc("/api/search", s.corsMiddleware(s.handleSearch))

	// Statistics
	mux.HandleFunc("/api/stats", s.corsMiddleware(s.handleStats))
	mux.HandleFunc("/api/stats/daily", s.corsMiddleware(s.handleDailyStats))

	// Camera
	mux.HandleFunc("/api/camera", s.corsMiddleware(s.handleCameraInfo))

	// Cache management
	mux.HandleFunc("/api/cache/clear", s.corsMiddleware(s.handleCacheClear))

	// ✅ НОВЫЕ ЭНДПОИНТЫ ДЛЯ АНАЛИТИКИ
	mux.HandleFunc("/api/analytics/summary", s.corsMiddleware(s.handleAnalyticsSummary))
	mux.HandleFunc("/api/analytics/detections/hourly", s.corsMiddleware(s.handleAnalyticsDetectionsHourly))
	mux.HandleFunc("/api/analytics/top-objects", s.corsMiddleware(s.handleAnalyticsTopObjects))
	mux.HandleFunc("/api/analytics/detections/count", s.corsMiddleware(s.handleAnalyticsDetectionsCount))

	// Static files (if needed)
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(s.outputDir))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем API endpoints
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Пропускаем /files/
		if strings.HasPrefix(r.URL.Path, "/files/") {
			http.NotFound(w, r)
			return
		}

		// ДЛЯ ВСЕХ остальных путей отдаем index.html
		http.ServeFile(w, r, "web/index.html")
	})
}

// CORS Middleware - расширенная версия
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow Svelte dev server
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// GetHealth godoc
// @Summary Health check
// @Description Проверка работоспособности API
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Router /health [get]
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	healthData := map[string]interface{}{
		"timestamp":  time.Now(),
		"version":    "1.0.0",
		"redis":      s.cache != nil,
		"clickhouse": s.analyticsClient != nil,
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "API server is running",
		Data:    healthData,
	})
}

// GetStatus godoc
// @Summary Get system status
// @Description Получить полный статус системы видеонаблюдения
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=StatusResponse}
// @Failure 500 {object} APIResponse
// @Router /status [get]
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Try to get from cache first
	if s.cache != nil {
		if cachedData, err := s.cache.GetSystemStatus(); err == nil {
			log.Println("✅ Cache HIT: system status")
			s.jsonResponse(w, http.StatusOK, APIResponse{
				Success: true,
				Data:    cachedData,
				Cached:  true,
			})
			return
		}
		log.Println("❌ Cache MISS: system status")
	}

	// Get from database
	camera, err := s.storage.GetCameraStatus()
	if err != nil {
		log.Printf("Error getting camera status: %v", err)
	}

	stats, err := s.storage.GetDatabaseStats()
	if err != nil {
		log.Printf("Error getting database stats: %v", err)
		stats = make(map[string]int)
	}

	storageUsage, err := s.storage.GetStorageUsage()
	if err != nil {
		log.Printf("Error getting storage usage: %v", err)
		storageUsage = 0
	} else {
		log.Printf("📊 Storage: %.2f MB", float64(storageUsage)/1024/1024)
	}

	systemStatus := "online"
	if camera != nil && !camera.IsOnline() {
		systemStatus = "camera_offline"
	}

	systemStats := SystemStats{
		Cameras:           stats["cameras"],
		Recordings:        stats["recordings"],
		Frames:            stats["frames"],
		Detections:        stats["detections"],
		Events:            stats["events"],
		StorageUsageBytes: storageUsage, // ✅ правильно типизировано
	}

	response := StatusResponse{
		SystemStatus: systemStatus,
		Camera:       camera,
		Stats:        systemStats,
		Uptime:       "N/A",
		LastUpdated:  time.Now(),
	}

	// Save to cache
	if s.cache != nil {
		if err := s.cache.SetSystemStatus(response); err != nil {
			log.Printf("Warning: failed to cache status: %v", err)
		}
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    response,
		Cached:  false,
	})
}

// GetRecordings godoc
// @Summary Get recordings list
// @Description Получить список видеозаписей
// @Tags recordings
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20) minimum(1) maximum(100)
// @Success 200 {object} APIResponse{data=[]RecordingResponse}
// @Failure 500 {object} APIResponse
// @Router /recordings [get]
func (s *Server) handleRecordings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	limit := s.getIntParam(r, "limit", 20)
	if limit > 100 {
		limit = 100
	}

	camera, _ := s.storage.GetCameraStatus()
	if camera == nil {
		s.jsonError(w, http.StatusInternalServerError, "Camera not found")
		return
	}

	// Try cache first
	if s.cache != nil {
		if cachedData, err := s.cache.GetCachedRecordings(camera.ID); err == nil {
			log.Println("✅ Cache HIT: recordings")
			s.jsonResponse(w, http.StatusOK, APIResponse{
				Success: true,
				Data:    cachedData,
				Cached:  true,
			})
			return
		}
		log.Println("❌ Cache MISS: recordings")
	}

	recordings, err := s.storage.GetRecordings(camera.ID, limit)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get recordings: %v", err))
		return
	}

	// Enrich recordings with file info
	enrichedRecordings := make([]RecordingResponse, 0, len(recordings))
	for _, rec := range recordings {
		enriched := RecordingResponse{
			Recording:   &rec,
			FileExists:  s.fileExists(rec.FilePath),
			DownloadURL: fmt.Sprintf("/api/download/%d", rec.ID),
			StreamURL:   fmt.Sprintf("/api/video/%d", rec.ID),
		}
		enrichedRecordings = append(enrichedRecordings, enriched)
	}

	// Save to cache
	if s.cache != nil {
		if err := s.cache.SetCachedRecordings(camera.ID, enrichedRecordings); err != nil {
			log.Printf("Warning: failed to cache recordings: %v", err)
		}
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    enrichedRecordings,
		Cached:  false,
	})
}

// GetRecordingByID godoc
// @Summary Get recording by ID
// @Description Получить информацию о конкретной записи
// @Tags recordings
// @Accept json
// @Produce json
// @Param id path int true "Recording ID"
// @Success 200 {object} APIResponse{data=RecordingResponse}
// @Failure 404 {object} APIResponse
// @Router /recordings/{id} [get]
func (s *Server) handleRecordingByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := s.extractIDFromPath(r.URL.Path, "/api/recordings/")
	if id == 0 {
		s.jsonError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	recording, err := s.storage.GetRecording(id)
	if err != nil {
		s.jsonError(w, http.StatusNotFound, "Recording not found")
		return
	}

	response := RecordingResponse{
		Recording:   recording,
		FileExists:  s.fileExists(recording.FilePath),
		DownloadURL: fmt.Sprintf("/api/download/%d", recording.ID),
		StreamURL:   fmt.Sprintf("/api/video/%d", recording.ID),
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    response,
	})
}

// GET /api/video/{id} - Stream video with range support
func (s *Server) handleVideoStream(w http.ResponseWriter, r *http.Request) {
	id := s.extractIDFromPath(r.URL.Path, "/api/video/")
	if id == 0 {
		s.jsonError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	recording, err := s.storage.GetRecording(id)
	if err != nil {
		s.jsonError(w, http.StatusNotFound, "Recording not found")
		return
	}

	if !s.fileExists(recording.FilePath) {
		s.jsonError(w, http.StatusNotFound, "Video file not found")
		return
	}

	s.serveVideoWithRange(w, r, recording.FilePath)
}

// GET /api/download/{id} - Download video file
func (s *Server) handleVideoDownload(w http.ResponseWriter, r *http.Request) {
	id := s.extractIDFromPath(r.URL.Path, "/api/download/")
	if id == 0 {
		s.jsonError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	recording, err := s.storage.GetRecording(id)
	if err != nil {
		s.jsonError(w, http.StatusNotFound, "Recording not found")
		return
	}

	if !s.fileExists(recording.FilePath) {
		s.jsonError(w, http.StatusNotFound, "Video file not found")
		return
	}

	// Set headers for download
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"recording_%d.mp4\"", recording.ID))

	http.ServeFile(w, r, recording.FilePath)
}

// GetFrames godoc
// @Summary Get frames list
// @Description Получить список кадров
// @Tags frames
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(50) minimum(1) maximum(200)
// @Param has_detection query bool false "Filter by detection"
// @Success 200 {object} APIResponse{data=[]FrameResponse}
// @Failure 500 {object} APIResponse
// @Router /frames [get]
func (s *Server) handleFrames(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	limit := s.getIntParam(r, "limit", 50)
	if limit > 200 {
		limit = 200
	}

	hasDetection := r.URL.Query().Get("has_detection")

	camera, _ := s.storage.GetCameraStatus()
	if camera == nil {
		s.jsonError(w, http.StatusInternalServerError, "Camera not found")
		return
	}

	var frames []storage.Frame
	var err error

	if hasDetection == "true" {
		frames, err = s.storage.GetFramesWithDetection(camera.ID, limit)
	} else {
		// Get recent frames
		end := time.Now()
		start := end.Add(-24 * time.Hour)
		frames, err = s.storage.GetFramesByTimeRange(camera.ID, start, end, limit)
	}

	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get frames: %v", err))
		return
	}

	// Enrich frames with URLs
	enrichedFrames := make([]FrameResponse, 0, len(frames))
	for _, frame := range frames {
		enriched := FrameResponse{
			Frame:        &frame,
			FileExists:   s.fileExists(frame.FilePath),
			ImageURL:     fmt.Sprintf("/api/image/%d", frame.ID),
			HasDetection: frame.HasDetection,
		}
		if frame.ThumbnailPath != nil {
			enriched.ThumbnailURL = fmt.Sprintf("/api/image/%d?thumbnail=true", frame.ID)
		}
		enrichedFrames = append(enrichedFrames, enriched)
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    enrichedFrames,
	})
}

// GET /api/frames/{id}
func (s *Server) handleFrameByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := s.extractIDFromPath(r.URL.Path, "/api/frames/")
	if id == 0 {
		s.jsonError(w, http.StatusBadRequest, "Invalid frame ID")
		return
	}

	frame, err := s.storage.GetFrame(id)
	if err != nil {
		s.jsonError(w, http.StatusNotFound, "Frame not found")
		return
	}

	response := FrameResponse{
		Frame:        frame,
		FileExists:   s.fileExists(frame.FilePath),
		ImageURL:     fmt.Sprintf("/api/image/%d", frame.ID),
		HasDetection: frame.HasDetection,
	}

	if frame.ThumbnailPath != nil {
		response.ThumbnailURL = fmt.Sprintf("/api/image/%d?thumbnail=true", frame.ID)
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    response,
	})
}

// GET /api/image/{id}?thumbnail=true
func (s *Server) handleImageServe(w http.ResponseWriter, r *http.Request) {
	id := s.extractIDFromPath(r.URL.Path, "/api/image/")
	if id == 0 {
		s.jsonError(w, http.StatusBadRequest, "Invalid frame ID")
		return
	}

	frame, err := s.storage.GetFrame(id)
	if err != nil {
		s.jsonError(w, http.StatusNotFound, "Frame not found")
		return
	}

	var filePath string
	useThumbnail := r.URL.Query().Get("thumbnail") == "true"

	if useThumbnail && frame.ThumbnailPath != nil {
		filePath = *frame.ThumbnailPath
	} else {
		filePath = frame.FilePath
	}

	if !s.fileExists(filePath) {
		s.jsonError(w, http.StatusNotFound, "Image file not found")
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	http.ServeFile(w, r, filePath)
}

// GetEvents godoc
// @Summary Get events list
// @Description Получить список событий системы
// @Tags events
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20) minimum(1) maximum(100)
// @Success 200 {object} APIResponse{data=[]storage.Event}
// @Failure 500 {object} APIResponse
// @Router /events [get]
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	limit := s.getIntParam(r, "limit", 20)
	if limit > 100 {
		limit = 100
	}

	events, err := s.storage.GetRecentEvents(limit)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get events: %v", err))
		return
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    events,
	})
}

// Search godoc
// @Summary Search events using Elasticsearch
// @Description Search events with full-text search and filters
// @Tags search
// @Accept json
// @Produce json
// @Param q query string false "Search query"
// @Param severity query string false "Filter by severity"
// @Param type query string false "Filter by event type"
// @Param limit query int false "Limit" default(20) minimum(1) maximum(100)
// @Success 200 {object} APIResponse{data=[]search.DetectionDocument}
// @Failure 500 {object} APIResponse
// @Failure 503 {object} APIResponse
// @Router /search [get]
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check if Elasticsearch is available
	if s.searchClient == nil {
		s.jsonError(w, http.StatusServiceUnavailable, "Search service not available")
		return
	}

	// Parse query parameters
	q := r.URL.Query().Get("q")
	severity := r.URL.Query().Get("severity")
	eventType := r.URL.Query().Get("type")
	limit := s.getIntParam(r, "limit", 20)
	if limit > 100 {
		limit = 100
	}

	// Build Elasticsearch query
	must := []interface{}{}

	// Add text search if provided
	if q != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  q,
				"fields": []string{"title", "message"},
			},
		})
	}

	// Add filters
	filters := []interface{}{}
	if severity != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"severity": severity,
			},
		})
	}
	if eventType != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"event_type": eventType,
			},
		})
	}

	// Construct query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   must,
				"filter": filters,
			},
		},
		"size": limit,
		"sort": []interface{}{
			map[string]interface{}{"timestamp": "desc"},
		},
	}

	// If no must clauses, match all
	if len(must) == 0 {
		query["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{"match_all": map[string]interface{}{}},
				},
				"filter": filters,
			},
		}
	}

	// Execute search
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	results, err := s.searchClient.Search(ctx, query)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, fmt.Sprintf("Search failed: %v", err))
		return
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    results,
	})
}

// GetStats godoc
// @Summary Get database statistics
// @Description Получить статистику базы данных
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]int}
// @Failure 500 {object} APIResponse
// @Router /stats [get]
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Try cache first
	if s.cache != nil {
		if cachedData, err := s.cache.GetStats(); err == nil {
			log.Println("✅ Cache HIT: stats")
			s.jsonResponse(w, http.StatusOK, APIResponse{
				Success: true,
				Data:    cachedData,
				Cached:  true,
			})
			return
		}
		log.Println("❌ Cache MISS: stats")
	}

	stats, err := s.storage.GetDatabaseStats()
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get stats: %v", err))
		return
	}

	// Save to cache
	if s.cache != nil {
		if err := s.cache.SetStats(stats); err != nil {
			log.Printf("Warning: failed to cache stats: %v", err)
		}
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    stats,
		Cached:  false,
	})
}

// GET /api/stats/daily
func (s *Server) handleDailyStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// TODO: Implement daily statistics
	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Daily stats not yet implemented",
		Data:    []interface{}{},
	})
}

// GetCameraInfo godoc
// @Summary Get camera information
// @Description Получить информацию о камере
// @Tags camera
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=storage.Camera}
// @Failure 500 {object} APIResponse
// @Router /camera [get]
func (s *Server) handleCameraInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	camera, err := s.storage.GetCameraStatus()
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get camera info: %v", err))
		return
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    camera,
	})
}

// ClearCache godoc
// @Summary Clear cache
// @Description Очистить весь кэш Redis
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 503 {object} APIResponse
// @Router /cache/clear [post]
func (s *Server) handleCacheClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if s.cache == nil {
		s.jsonError(w, http.StatusServiceUnavailable, "Cache service not available")
		return
	}

	if err := s.cache.InvalidateAll(); err != nil {
		s.jsonError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to clear cache: %v", err))
		return
	}

	log.Println("🗑️  Cache cleared")

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Cache cleared successfully",
	})
}

// ✅ НОВЫЕ ОБРАБОТЧИКИ ДЛЯ АНАЛИТИКИ

// GetAnalyticsSummary godoc
// @Summary Get analytics summary
// @Description Получить сводку аналитики из ClickHouse
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=AnalyticsSummaryResponse}
// @Failure 500 {object} APIResponse
// @Router /analytics/summary [get]
func (s *Server) handleAnalyticsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if s.analyticsClient == nil {
		s.jsonResponse(w, http.StatusOK, APIResponse{
			Success: true,
			Data: AnalyticsSummaryResponse{
				ClickHouseAvailable: false,
			},
			Message: "ClickHouse analytics not available",
		})
		return
	}

	camera, _ := s.storage.GetCameraStatus()
	cameraID := uint32(1)
	if camera != nil {
		cameraID = uint32(camera.ID)
	}

	// Получить общее количество детекций
	totalDetections, err := s.analyticsClient.GetTotalDetectionsCount(cameraID)
	if err != nil {
		log.Printf("Warning: failed to get total detections: %v", err)
		totalDetections = 0
	}

	// Получить количество детекций за сегодня
	detectionsToday, err := s.analyticsClient.GetDetectionsCountToday(cameraID)
	if err != nil {
		log.Printf("Warning: failed to get today's detections: %v", err)
		detectionsToday = 0
	}

	// Получить топ обнаруженных объектов (за последние 7 дней)
	topObjects, err := s.analyticsClient.GetTopDetectedObjects(cameraID, 7, 10)
	if err != nil {
		log.Printf("Warning: failed to get top objects: %v", err)
		topObjects = []analytics.TopDetectedObjects{}
	}

	// Преобразовать в response структуру
	topObjectsResponse := make([]TopObjectsResponse, 0, len(topObjects))
	for _, obj := range topObjects {
		topObjectsResponse = append(topObjectsResponse, TopObjectsResponse{
			ObjectClass:     obj.ObjectClass,
			TotalDetections: obj.TotalDetections,
			AvgConfidence:   obj.AvgConfidence,
		})
	}

	// Получить детекции по часам (за последние 24 часа)
	detectionsByHour, err := s.analyticsClient.GetDetectionsByHour(cameraID, 24)
	if err != nil {
		log.Printf("Warning: failed to get detections by hour: %v", err)
		detectionsByHour = []analytics.DetectionStats{}
	}

	// Преобразовать в response структуру
	hourlyResponse := make([]DetectionsHourlyResponse, 0, len(detectionsByHour))
	for _, stat := range detectionsByHour {
		hourlyResponse = append(hourlyResponse, DetectionsHourlyResponse{
			Hour:            stat.Hour.Format("2006-01-02 15:00"),
			ObjectClass:     stat.ObjectClass,
			TotalDetections: stat.TotalDetections,
			AvgConfidence:   stat.AvgConfidence,
			MinConfidence:   stat.MinConfidence,
			MaxConfidence:   stat.MaxConfidence,
		})
	}

	response := AnalyticsSummaryResponse{
		TotalDetections:     totalDetections,
		DetectionsToday:     detectionsToday,
		TopObjects:          topObjectsResponse,
		DetectionsByHour:    hourlyResponse,
		ClickHouseAvailable: true,
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetAnalyticsDetectionsHourly godoc
// @Summary Get detections by hour
// @Description Получить детекции по часам
// @Tags analytics
// @Accept json
// @Produce json
// @Param hours query int false "Hours to look back" default(24) minimum(1) maximum(168)
// @Success 200 {object} APIResponse{data=[]DetectionsHourlyResponse}
// @Failure 500 {object} APIResponse
// @Router /analytics/detections/hourly [get]
func (s *Server) handleAnalyticsDetectionsHourly(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if s.analyticsClient == nil {
		s.jsonError(w, http.StatusServiceUnavailable, "ClickHouse analytics not available")
		return
	}

	hours := s.getIntParam(r, "hours", 24)
	if hours > 168 { // max 7 days
		hours = 168
	}

	camera, _ := s.storage.GetCameraStatus()
	cameraID := uint32(1)
	if camera != nil {
		cameraID = uint32(camera.ID)
	}

	detections, err := s.analyticsClient.GetDetectionsByHour(cameraID, hours)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get detections: %v", err))
		return
	}

	// Преобразовать в response структуру
	response := make([]DetectionsHourlyResponse, 0, len(detections))
	for _, stat := range detections {
		response = append(response, DetectionsHourlyResponse{
			Hour:            stat.Hour.Format("2006-01-02 15:00"),
			ObjectClass:     stat.ObjectClass,
			TotalDetections: stat.TotalDetections,
			AvgConfidence:   stat.AvgConfidence,
			MinConfidence:   stat.MinConfidence,
			MaxConfidence:   stat.MaxConfidence,
		})
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetAnalyticsTopObjects godoc
// @Summary Get top detected objects
// @Description Получить топ обнаруженных объектов
// @Tags analytics
// @Accept json
// @Produce json
// @Param days query int false "Days to look back" default(7) minimum(1) maximum(90)
// @Param limit query int false "Limit" default(10) minimum(1) maximum(50)
// @Success 200 {object} APIResponse{data=[]TopObjectsResponse}
// @Failure 500 {object} APIResponse
// @Router /analytics/top-objects [get]
func (s *Server) handleAnalyticsTopObjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if s.analyticsClient == nil {
		s.jsonError(w, http.StatusServiceUnavailable, "ClickHouse analytics not available")
		return
	}

	days := s.getIntParam(r, "days", 7)
	limit := s.getIntParam(r, "limit", 10)

	if days > 90 {
		days = 90
	}
	if limit > 50 {
		limit = 50
	}

	camera, _ := s.storage.GetCameraStatus()
	cameraID := uint32(1)
	if camera != nil {
		cameraID = uint32(camera.ID)
	}

	topObjects, err := s.analyticsClient.GetTopDetectedObjects(cameraID, days, limit)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get top objects: %v", err))
		return
	}

	// Преобразовать в response структуру
	response := make([]TopObjectsResponse, 0, len(topObjects))
	for _, obj := range topObjects {
		response = append(response, TopObjectsResponse{
			ObjectClass:     obj.ObjectClass,
			TotalDetections: obj.TotalDetections,
			AvgConfidence:   obj.AvgConfidence,
		})
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    response,
	})
}

// GET /api/analytics/detections/count
func (s *Server) handleAnalyticsDetectionsCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if s.analyticsClient == nil {
		s.jsonError(w, http.StatusServiceUnavailable, "ClickHouse analytics not available")
		return
	}

	camera, _ := s.storage.GetCameraStatus()
	cameraID := uint32(1)
	if camera != nil {
		cameraID = uint32(camera.ID)
	}

	totalCount, err := s.analyticsClient.GetTotalDetectionsCount(cameraID)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get count: %v", err))
		return
	}

	todayCount, err := s.analyticsClient.GetDetectionsCountToday(cameraID)
	if err != nil {
		log.Printf("Warning: failed to get today's count: %v", err)
		todayCount = 0
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]uint64{
			"total": totalCount,
			"today": todayCount,
		},
	})
}

// Helper methods

func (s *Server) serveVideoWithRange(w http.ResponseWriter, r *http.Request, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to open video file")
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to get file info")
		return
	}

	size := stat.Size()
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")

	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		w.WriteHeader(http.StatusOK)
		io.Copy(w, file)
		return
	}

	// Parse range header
	ranges := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
	if len(ranges) != 2 {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	start, _ := strconv.ParseInt(ranges[0], 10, 64)
	end := size - 1
	if ranges[1] != "" {
		end, _ = strconv.ParseInt(ranges[1], 10, 64)
	}

	if start > end || start < 0 || end >= size {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	contentLength := end - start + 1
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.WriteHeader(http.StatusPartialContent)

	file.Seek(start, 0)
	io.CopyN(w, file, contentLength)
}

func (s *Server) jsonResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) jsonError(w http.ResponseWriter, statusCode int, message string) {
	s.jsonResponse(w, statusCode, APIResponse{
		Success: false,
		Error:   message,
	})
}

func (s *Server) getIntParam(r *http.Request, param string, defaultValue int) int {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

func (s *Server) extractIDFromPath(path, prefix string) int {
	idStr := strings.TrimPrefix(path, prefix)
	idStr = strings.Split(idStr, "/")[0]
	id, _ := strconv.Atoi(idStr)
	return id
}

func (s *Server) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
