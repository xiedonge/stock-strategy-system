package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/xiedonge/stock-strategy-system/backend/internal/config"
	"github.com/xiedonge/stock-strategy-system/backend/internal/db"
	"github.com/xiedonge/stock-strategy-system/backend/internal/handlers"
	"github.com/xiedonge/stock-strategy-system/backend/internal/models"
	"github.com/xiedonge/stock-strategy-system/backend/internal/services"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	stockService := services.NewStockService(database)
	strategyService := services.NewStrategyService(database)
	analysisService := services.NewAnalysisService(database, stockService, strategyService)
	syncService := initSyncService(cfg.DBPath)

	ensureDefaultStrategy(strategyService)

	router := handlers.NewRouter(stockService, strategyService, analysisService, syncService)

	log.Printf("stock strategy backend listening on :%s", cfg.Port)
	if err := router.Run("0.0.0.0:" + cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func ensureDefaultStrategy(strategyService *services.StrategyService) {
	strategies, err := strategyService.List()
	if err != nil {
		log.Printf("skip default strategy: %v", err)
		return
	}
	if len(strategies) > 0 {
		return
	}

	defaultStrategy := &models.Strategy{
		Name:        "MA交叉默认策略",
		Description: "短期均线向上穿越长期均线时入选",
		Type:        "ma_crossover",
		ParamsJSON:  `{"short_window":5,"long_window":20}`,
	}
	if err := strategyService.Create(defaultStrategy); err != nil {
		log.Printf("failed to create default strategy: %v", err)
	}
}

func initSyncService(dbPath string) *services.SyncService {
	scriptPath := os.Getenv("AKSHARE_SYNC_SCRIPT")
	if scriptPath == "" {
		scriptPath = filepath.Join("..", "scripts", "sync_akshare.sh")
	}
	absScript, err := filepath.Abs(scriptPath)
	if err != nil {
		absScript = scriptPath
	}
	if _, err := os.Stat(absScript); err != nil {
		alt := filepath.Join("scripts", "sync_akshare.sh")
		if altAbs, err := filepath.Abs(alt); err == nil {
			absScript = altAbs
		}
	}

	timeoutSeconds := 300
	if raw := os.Getenv("SYNC_TIMEOUT_SECONDS"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			timeoutSeconds = parsed
		}
	}

	return services.NewSyncService(absScript, dbPath, time.Duration(timeoutSeconds)*time.Second)
}
