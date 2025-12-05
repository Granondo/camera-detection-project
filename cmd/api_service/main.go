// @title Camera Surveillance API
// @version 1.0
// @description API для системы видеонаблюдения с YOLO детекцией
// @host localhost:8080
// @BasePath /api
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"camera-detection-project/internal/analytics"
	"camera-detection-project/internal/api"
	"camera-detection-project/internal/cache"
	"camera-detection-project/internal/config"
	"camera-detection-project/internal/search"
	"camera-detection-project/internal/storage"

	_ "camera-detection-project/docs" // swagger docs
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	log.Println("🚀 Starting Camera Detection API Server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// ✅ Initialize Redis
	redisClient, err := cache.NewRedisClient(&cache.RedisConfig{
		Host:     cfg.RedisHost,
		Port:     cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
		TTL:      cfg.RedisCacheTTL,
	})

	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	cacheService := cache.NewCacheService(redisClient)

	// ✅ Initialize PostgreSQL storage service
	storageService, err := storage.NewService(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize storage service: %v", err)
	}
	defer storageService.Close()

	log.Println("✅ Connected to PostgreSQL database")

	// Initialize database tables if needed
	if err := storageService.InitializeDatabase(); err != nil {
		log.Printf("⚠️  Warning: Could not initialize database: %v", err)
	}

	// ✅ Initialize ClickHouse analytics
	var clickhouseClient *analytics.ClickHouseClient
	clickhouseClient, err = analytics.NewClickHouseClient(&analytics.ClickHouseConfig{
		Host:     cfg.ClickHouseHost,
		Port:     cfg.ClickHousePort,
		Database: cfg.ClickHouseDatabase,
		Username: cfg.ClickHouseUsername,
		Password: cfg.ClickHousePassword,
		Debug:    cfg.ClickHouseDebug,
	})

	if err != nil {
		log.Printf("⚠️  Warning: Failed to connect to ClickHouse: %v", err)
		log.Println("⚠️  Continuing without analytics...")
		clickhouseClient = nil
	} else {
		defer clickhouseClient.Close()
		log.Println("✅ Connected to ClickHouse database")

		// Автоматически создать таблицы
		if err := clickhouseClient.Migrate(); err != nil {
			log.Printf("⚠️  Warning: ClickHouse migration failed: %v", err)
		} else {
			log.Println("✅ ClickHouse tables ready")
		}
	}

	// ✅ Initialize Elasticsearch
	var searchClient *search.ElasticsearchClient
	if cfg.ElasticsearchEnabled {
		searchClient, err = search.NewElasticsearchClient(&search.Config{
			Addresses: []string{cfg.ElasticsearchURL},
			Index:     cfg.ElasticsearchIndex,
		})

		if err != nil {
			log.Printf("⚠️  Warning: Failed to connect to Elasticsearch: %v", err)
			log.Println("⚠️  Continuing without Elasticsearch (no search indexing)...")
			searchClient = nil
		} else {
			log.Println("✅ Connected to Elasticsearch")
		}
	}

	// ✅ Create API server with all services
	apiServer := api.NewServerWithAnalytics(storageService, cacheService, clickhouseClient, searchClient, cfg.OutputDir)

	// Log enabled services
	var services []string
	if clickhouseClient != nil {
		services = append(services, "ClickHouse analytics")
	}
	if searchClient != nil {
		services = append(services, "Elasticsearch search")
	}
	if len(services) > 0 {
		log.Printf("✅ API Server initialized with: %v", services)
	} else {
		log.Println("✅ API Server initialized (basic mode)")
	}

	// Setup routes
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	// Add logging middleware
	handler := loggingMiddleware(mux)

	// Get port from environment or use default
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🌐 API Server listening on http://localhost:%s", port)
		log.Println("📋 Available endpoints:")
		log.Println("   GET  /api/health           - Health check")
		log.Println("   GET  /api/status           - System status")
		log.Println("   GET  /api/recordings       - List recordings")
		log.Println("   GET  /api/recordings/{id}  - Get recording details")
		log.Println("   GET  /api/video/{id}       - Stream video (with range support)")
		log.Println("   GET  /api/download/{id}    - Download video")
		log.Println("   GET  /api/frames           - List frames")
		log.Println("   GET  /api/frames/{id}      - Get frame details")
		log.Println("   GET  /api/image/{id}       - Serve frame image")
		log.Println("   GET  /api/events           - List events")
		log.Println("   GET  /api/stats            - Database statistics")
		log.Println("   GET  /api/camera           - Camera information")

		if searchClient != nil {
			log.Println("")
			log.Println("🔍 Search endpoints (Elasticsearch):")
			log.Println("   GET  /api/search           - Search events (q, severity, type, limit)")
		}

		if clickhouseClient != nil {
			log.Println("")
			log.Println("📊 Analytics endpoints (ClickHouse):")
			log.Println("   GET  /api/analytics/summary              - Analytics summary")
			log.Println("   GET  /api/analytics/detections/hourly    - Detections by hour")
			log.Println("   GET  /api/analytics/top-objects          - Top detected objects")
			log.Println("   GET  /api/analytics/detections/count     - Total detections count")
		}
		log.Println("")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down API server...")

	// Gracefully shutdown the server
	if err := server.Close(); err != nil {
		log.Printf("❌ Error during server shutdown: %v", err)
	}

	log.Println("✅ API Server stopped")
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
