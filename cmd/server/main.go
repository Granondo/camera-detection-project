package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"camera-detection-project/internal/camera"
	"camera-detection-project/internal/config"
	"camera-detection-project/internal/storage"
)

func main() {
	log.Println("Starting camera detection service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage service
	storageService, err := storage.NewService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage service: %v", err)
	}
	defer storageService.Close()

	// Initialize database tables if needed
	if err := storageService.InitializeDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create system startup event
	if err := storageService.CreateSystemEvent(
		storage.EventTypeSystemStart,
		storage.SeverityLow,
		"System Started",
		"Camera detection system has been started",
	); err != nil {
		log.Printf("Warning: Could not create startup event: %v", err)
	}

	// Create camera client with storage integration
	client, err := camera.NewFFmpegClientWithStorage(cfg, storageService)
	if err != nil {
		log.Fatalf("Failed to create camera client: %v", err)
	}
	defer client.Close()

	// Update camera status to active
	if err := storageService.UpdateCameraStatus(storage.CameraStatusActive); err != nil {
		log.Printf("Warning: Could not update camera status: %v", err)
	}

	// Start video processing
	if err := client.Start(); err != nil {
		log.Fatalf("Failed to start camera client: %v", err)
	}

	log.Println("Camera detection service started successfully")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")

	// Update camera status to inactive
	if err := storageService.UpdateCameraStatus(storage.CameraStatusInactive); err != nil {
		log.Printf("Warning: Could not update camera status: %v", err)
	}

	// Create system shutdown event
	if err := storageService.CreateSystemEvent(
		storage.EventTypeSystemStop,
		storage.SeverityLow,
		"System Stopped",
		"Camera detection system has been stopped",
	); err != nil {
		log.Printf("Warning: Could not create shutdown event: %v", err)
	}

	// Stop camera client
	client.Stop()

	log.Println("Shutdown completed")
}