package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xiedonge/stock-strategy-system/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Open opens a SQLite database and applies schema migrations.
func Open(path string) (*gorm.DB, error) {
	if err := ensureDir(path); err != nil {
		return nil, err
	}

	database, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// AutoMigrate keeps the schema aligned with the models.
	if err := database.AutoMigrate(
		&models.Stock{},
		&models.KLine{},
		&models.Strategy{},
		&models.Backtest{},
		&models.BacktestPoint{},
	); err != nil {
		return nil, err
	}

	return database, nil
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." {
		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create db dir: %w", err)
	}
	return nil
}
