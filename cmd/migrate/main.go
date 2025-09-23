package main

import (
	"log"
	"camera-detection-project/internal/config"
	"camera-detection-project/internal/storage"
)

func main() {
	log.Println("🗄️  Database Migration Utility")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create database configuration
	dbConfig := &storage.DatabaseConfig{
		Host:         cfg.DatabaseHost,
		Port:         cfg.DatabasePort,
		User:         cfg.DatabaseUser,
		Password:     cfg.DatabasePassword,
		Database:     cfg.DatabaseName,
		SSLMode:      cfg.DatabaseSSLMode,
		MaxOpenConns: 10,
		MaxIdleConns: 2,
	}

	// Connect to database
	db, err := storage.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Connected to database successfully")

	// Create tables
	if err := db.CreateTables(); err != nil {
		log.Fatalf("❌ Failed to create tables: %v", err)
	}

	log.Println("✅ Database tables created successfully")

	// Show statistics
	stats, err := db.GetDatabaseStats()
	if err != nil {
		log.Printf("Warning: Could not get database stats: %v", err)
	} else {
		log.Println("📊 Database Statistics:")
		for table, count := range stats {
			log.Printf("  %s: %d records", table, count)
		}
	}

	log.Println("🎉 Migration completed successfully!")
}