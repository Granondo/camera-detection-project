package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"camera-detection-project/internal/camera"
	"camera-detection-project/internal/config"
)

func main() {
	log.Println("Starting camera detection service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create camera client
	client, err := camera.NewFFmpegClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create camera client: %v", err)
	}
	defer client.Close()

	// Start video processing
	if err := client.Start(); err != nil {
		log.Fatalf("Failed to start camera client: %v", err)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	client.Stop()
}