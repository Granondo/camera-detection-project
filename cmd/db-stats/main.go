package main

import (
	"fmt"
	"log"

	"camera-detection-project/internal/config"
	"camera-detection-project/internal/storage"
)

func main() {
	log.Println("📊 Database Statistics Utility")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage service
	storageService, err := storage.NewService(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer storageService.Close()

	log.Println("✅ Connected to database successfully")
	fmt.Println()

	// Get basic database statistics
	stats, err := storageService.GetDatabaseStats()
	if err != nil {
		log.Fatalf("❌ Failed to get database statistics: %v", err)
	}

	// Display statistics
	fmt.Println("🗄️  DATABASE STATISTICS")
	fmt.Println("========================")
	fmt.Printf("📹 Cameras:      %d\n", stats["cameras"])
	fmt.Printf("🎬 Recordings:   %d\n", stats["recordings"])
	fmt.Printf("🖼️  Frames:       %d\n", stats["frames"])
	fmt.Printf("🎯 Detections:   %d\n", stats["detections"])
	fmt.Printf("📢 Events:       %d\n", stats["events"])
	fmt.Println()

	// Get camera status
	camera, err := storageService.GetCameraStatus()
	if err != nil {
		log.Printf("⚠️  Warning: Could not get camera status: %v", err)
	} else {
		fmt.Println("📹 CAMERA STATUS")
		fmt.Println("=================")
		fmt.Printf("Name:        %s\n", camera.Name)
		fmt.Printf("Status:      %s", camera.Status)
		if camera.IsOnline() {
			fmt.Printf(" 🟢 (Online)")
		} else {
			fmt.Printf(" 🔴 (Offline)")
		}
		fmt.Println()
		if camera.LastPing != nil {
			fmt.Printf("Last Ping:   %s\n", camera.LastPing.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("Last Ping:   Never\n")
		}
		fmt.Printf("Created:     %s\n", camera.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	// Get recent events
	events, err := storageService.GetRecentEvents(5)
	if err != nil {
		log.Printf("⚠️  Warning: Could not get recent events: %v", err)
	} else if len(events) > 0 {
		fmt.Println("📢 RECENT EVENTS (Last 5)")
		fmt.Println("==========================")
		for _, event := range events {
			severityIcon := getSeverityIcon(event.Severity)
			fmt.Printf("%s %s - %s\n", severityIcon, event.Timestamp.Format("15:04:05"), event.Title)
			if event.Message != "" {
				fmt.Printf("    %s\n", event.Message)
			}
		}
		fmt.Println()
	}

	// Get storage usage
	storageUsage, err := storageService.GetStorageUsage()
	if err != nil {
		log.Printf("⚠️  Warning: Could not get storage usage: %v", err)
	} else {
		fmt.Println("💾 STORAGE USAGE")
		fmt.Println("=================")
		fmt.Printf("Total Size:  %s\n", formatBytes(storageUsage))
		fmt.Printf("Directory:   %s\n", cfg.OutputDir)
		fmt.Println()
	}

	fmt.Println("✅ Statistics retrieved successfully!")
}

func getSeverityIcon(severity string) string {
	switch severity {
	case storage.SeverityCritical:
		return "🚨"
	case storage.SeverityHigh:
		return "⚠️"
	case storage.SeverityMedium:
		return "ℹ️"
	case storage.SeverityLow:
		return "✅"
	default:
		return "📝"
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(bytes)/float64(div), "KMGTPE"[exp])
}