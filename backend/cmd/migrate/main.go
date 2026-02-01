package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"massrouter.ai/backend/internal/config"
	"massrouter.ai/backend/pkg/database"

	"gorm.io/gorm"
)

type Migration struct {
	gorm.Model
	Name  string `gorm:"size:255;not null"`
	Batch int
}

var (
	command = flag.String("command", "", "Migration command: up, down, create")
	name    = flag.String("name", "", "Migration name (for create command)")
)

func main() {
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.NewPostgresDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	switch *command {
	case "up":
		if err := runMigrationsUp(db.GetDB()); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("✅ Migrations applied successfully")
	case "down":
		if err := runMigrationsDown(db.GetDB()); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("✅ Migrations rolled back successfully")
	case "create":
		if *name == "" {
			log.Fatal("Migration name is required for create command")
		}
		if err := createMigration(*name); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
		fmt.Println("✅ Migration files created successfully")
	default:
		log.Fatal("Please specify a command: up, down, or create")
	}
}

func runMigrationsUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrationsDir := "./migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	appliedMigrations := make(map[string]bool)
	var applied []Migration
	db.Find(&applied)
	for _, m := range applied {
		appliedMigrations[m.Name] = true
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		if appliedMigrations[file.Name()] {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to begin transaction: %w", tx.Error)
		}

		if err := tx.Exec(string(content)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
		}

		migration := Migration{
			Name:  file.Name(),
			Batch: 1,
		}
		if err := tx.Create(&migration).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", file.Name(), err)
		}

		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		fmt.Printf("✅ Applied migration: %s\n", file.Name())
	}

	return nil
}

func runMigrationsDown(db *gorm.DB) error {
	var lastBatch []Migration
	if err := db.Where("batch = (SELECT MAX(batch) FROM migrations)").Find(&lastBatch).Error; err != nil {
		return fmt.Errorf("failed to find last batch: %w", err)
	}

	if len(lastBatch) == 0 {
		fmt.Println("No migrations to roll back")
		return nil
	}

	for _, migration := range lastBatch {
		if !strings.HasSuffix(migration.Name, ".up.sql") {
			continue
		}
		downFileName := strings.TrimSuffix(migration.Name, ".up.sql") + ".down.sql"
		downFile := filepath.Join("./migrations", downFileName)
		content, err := os.ReadFile(downFile)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", downFileName, err)
		}

		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to begin transaction: %w", tx.Error)
		}

		if err := tx.Exec(string(content)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute rollback %s: %w", downFileName, err)
		}

		if err := tx.Delete(&migration).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete migration record %s: %w", migration.Name, err)
		}

		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		fmt.Printf("✅ Rolled back migration: %s\n", downFileName)
	}

	return nil
}

func createMigration(name string) error {
	timestamp := time.Now().Format("20060102150405")
	upFileName := fmt.Sprintf("%s_%s.up.sql", timestamp, name)
	downFileName := fmt.Sprintf("%s_%s.down.sql", timestamp, name)

	upContent := fmt.Sprintf("-- Migration up: %s\n", name)
	downContent := fmt.Sprintf("-- Migration down: %s\n", name)

	if err := os.WriteFile(filepath.Join("./migrations", upFileName), []byte(upContent), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	if err := os.WriteFile(filepath.Join("./migrations", downFileName), []byte(downContent), 0644); err != nil {
		os.Remove(filepath.Join("./migrations", upFileName))
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	fmt.Printf("Created migration files:\n")
	fmt.Printf("  Up: %s\n", upFileName)
	fmt.Printf("  Down: %s\n", downFileName)

	return nil
}
