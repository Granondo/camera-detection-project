package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"camera-detection-project/internal/analytics"
	"camera-detection-project/internal/camera"
	"camera-detection-project/internal/config"
	"camera-detection-project/internal/search"
	"camera-detection-project/internal/storage"
	"camera-detection-project/internal/queue"
)

func main() {
	log.Println("🚀 Starting camera detection service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Initialize PostgreSQL storage
	storageService, err := storage.NewService(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize storage service: %v", err)
	}
	defer storageService.Close()

	// Initialize database tables if needed
	if err := storageService.InitializeDatabase(); err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}

	// ✨ Initialize ClickHouse (ДОБАВЬ ЭТО)
	clickhouseClient, err := analytics.NewClickHouseClient(&analytics.ClickHouseConfig{
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
		
		// Автоматически создать таблицы
		if err := clickhouseClient.Migrate(); err != nil {
			log.Printf("⚠️  Warning: ClickHouse migration failed: %v", err)
		}
	}

	var queueClient *queue.RabbitMQClient
	if cfg.RabbitMQEnabled {
		queueClient, err = queue.NewRabbitMQClient(&queue.RabbitMQConfig{
			Host:     cfg.RabbitMQHost,
			Port:     cfg.RabbitMQPort,
			User:     cfg.RabbitMQUser,
			Password: cfg.RabbitMQPassword,
			VHost:    cfg.RabbitMQVHost,
		})

		if err != nil {
			log.Printf("⚠️  Warning: Failed to connect to RabbitMQ: %v", err)
			log.Println("⚠️  Continuing without RabbitMQ (sync detection mode)...")
			queueClient = nil
		} else {
			defer queueClient.Close()
			log.Println("✅ Connected to RabbitMQ")
		}
	}

	// ✨ Initialize Elasticsearch
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

	// Create system startup event
	_, err = storageService.CreateSystemEvent(
		storage.EventTypeSystemStart,
		storage.SeverityLow,
		"System Started",
		"Camera detection system has been started",
	)
	if err != nil {
		log.Printf("⚠️  Warning: Could not create startup event: %v", err)
	}

	// Log to ClickHouse if available
	if clickhouseClient != nil {
		event := &analytics.SystemEvent{
			Timestamp:  time.Now(),
			EventType:  "system_start",
			Severity:   "low",
			Service:    "camera-service",
			Title:      "System Started",
			Message:    "Camera detection system has been started",
			CameraID:   nil,
			Metadata:   "{}",
		}
		if err := clickhouseClient.InsertSystemEvent(event); err != nil {
			log.Printf("⚠️  Warning: Could not log to ClickHouse: %v", err)
		}
	}

	// Create camera client with storage, analytics, and search
	client, err := camera.NewFFmpegClientWithStorageAndAnalytics(cfg, storageService, clickhouseClient, queueClient, searchClient)
	if err != nil {
		log.Fatalf("❌ Failed to create camera client: %v", err)
	}
	defer client.Close()

	// Update camera status to active
	if err := storageService.UpdateCameraStatus(storage.CameraStatusActive); err != nil {
		log.Printf("⚠️  Warning: Could not update camera status: %v", err)
	}

	// Ping camera status every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			if err := storageService.UpdateCameraPing(); err != nil {
				log.Printf("⚠️  Warning: Could not update camera ping: %v", err)
			}
		}
	}()

	// Start video processing
	if err := client.Start(); err != nil {
		log.Fatalf("❌ Failed to start camera client: %v", err)
	}

	log.Println("✅ Camera detection service started successfully")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("🛑 Shutting down...")

	// Update camera status to inactive
	if err := storageService.UpdateCameraStatus(storage.CameraStatusInactive); err != nil {
		log.Printf("⚠️  Warning: Could not update camera status: %v", err)
	}

	// Create system shutdown event
	_, err = storageService.CreateSystemEvent(
		storage.EventTypeSystemStop,
		storage.SeverityLow,
		"System Stopped",
		"Camera detection system has been stopped",
	)
	if err != nil {
		log.Printf("⚠️  Warning: Could not create shutdown event: %v", err)
	}

	// Log to ClickHouse
	if clickhouseClient != nil {
		event := &analytics.SystemEvent{
			Timestamp:  time.Now(),
			EventType:  "system_stop",
			Severity:   "low",
			Service:    "camera-service",
			Title:      "System Stopped",
			Message:    "Camera detection system has been stopped",
			CameraID:   nil,
			Metadata:   "{}",
		}
		if err := clickhouseClient.InsertSystemEvent(event); err != nil {
			log.Printf("⚠️  Warning: Could not log to ClickHouse: %v", err)
		}
	}

	// Stop camera client
	client.Stop()

	log.Println("✅ Shutdown completed")
}