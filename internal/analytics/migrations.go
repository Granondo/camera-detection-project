package analytics

import (
	"os"
	"fmt"
	"log"
)

// Migrate создает все таблицы если их нет
// Migrate создает все таблицы если их нет
func (c *ClickHouseClient) Migrate() error {
	log.Println("🔄 Checking ClickHouse tables...")

	// Проверить существуют ли таблицы
	exists, err := c.CheckTablesExist()
	if err != nil {
		return fmt.Errorf("failed to check tables: %w", err)
	}

	if exists {
		log.Println("✅ ClickHouse tables already exist")
		return nil
	}

	// Если таблиц нет, попробовать создать через SQL скрипты
	log.Println("⚠️  Tables don't exist, creating from init scripts...")
	
	// Читаем SQL файл
	sqlBytes, err := os.ReadFile("/docker-entrypoint-initdb.d/01_create_tables.sql")
	if err != nil {
		log.Printf("⚠️  Could not read init SQL: %v", err)
		log.Println("📝 Tables should be created by Docker init scripts")
		return nil // не фатально, таблицы могут быть уже созданы
	}

	// Выполняем SQL
	if err := c.conn.Exec(c.ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("failed to execute init SQL: %w", err)
	}

	log.Println("✅ ClickHouse tables created successfully")
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