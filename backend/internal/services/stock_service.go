package services

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/xiedonge/stock-strategy-system/backend/internal/models"
	"gorm.io/gorm"
)

// StockService encapsulates stock and kline persistence.
type StockService struct {
	db *gorm.DB
}

// NewStockService constructs a StockService.
func NewStockService(db *gorm.DB) *StockService {
	return &StockService{db: db}
}

// ListStocks returns all known stocks.
func (s *StockService) ListStocks() ([]models.Stock, error) {
	var stocks []models.Stock
	if err := s.db.Order("code asc").Find(&stocks).Error; err != nil {
		return nil, err
	}
	return stocks, nil
}

// GetKLines fetches klines for a stock code and interval.
func (s *StockService) GetKLines(code, interval string, limit int) ([]models.KLine, error) {
	if interval == "" {
		interval = "1d"
	}
	if limit <= 0 || limit > 2000 {
		limit = 500
	}

	var klines []models.KLine
	query := s.db.Where("stock_code = ? AND interval = ?", code, interval).Order("time asc").Limit(limit)
	if err := query.Find(&klines).Error; err != nil {
		return nil, err
	}
	return klines, nil
}

// UpsertStock ensures a stock row exists.
func (s *StockService) UpsertStock(stock models.Stock) error {
	var existing models.Stock
	if err := s.db.Where("code = ?", stock.Code).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.db.Create(&stock).Error
		}
		return err
	}

	existing.Name = stock.Name
	existing.Exchange = stock.Exchange
	return s.db.Save(&existing).Error
}

// SaveKLines persists kline entries.
func (s *StockService) SaveKLines(klines []models.KLine) error {
	if len(klines) == 0 {
		return nil
	}
	return s.db.Create(&klines).Error
}

// SeedDemoData populates the database with deterministic sample data.
func (s *StockService) SeedDemoData() error {
	seed := rand.New(rand.NewSource(42))
	stocks := []models.Stock{
		{Code: "600519", Name: "贵州茅台", Exchange: "SH"},
		{Code: "000001", Name: "平安银行", Exchange: "SZ"},
	}
	for _, stock := range stocks {
		if err := s.UpsertStock(stock); err != nil {
			return err
		}
		if err := s.seedKLines(stock.Code, "1d", 160, seed, 120.0, 0.012); err != nil {
			return err
		}
		if err := s.seedKLines(stock.Code, "30m", 240, seed, 120.0, 0.006); err != nil {
			return err
		}
	}
	return nil
}

func (s *StockService) seedKLines(code, interval string, points int, seed *rand.Rand, base, volatility float64) error {
	if points <= 0 {
		return nil
	}

	start := time.Now().AddDate(0, 0, -points)
	price := base
	klines := make([]models.KLine, 0, points)

	for i := 0; i < points; i++ {
		change := (seed.Float64()*2 - 1) * volatility
		price = math.Max(5, price*(1+change))
		open := price * (1 + (seed.Float64()-0.5)*0.002)
		close := price
		high := math.Max(open, close) * (1 + seed.Float64()*0.003)
		low := math.Min(open, close) * (1 - seed.Float64()*0.003)
		volume := 100000 + seed.Float64()*200000

		stamp := start.AddDate(0, 0, i)
		if interval == "30m" {
			stamp = start.Add(time.Duration(i) * 30 * time.Minute)
		}

		klines = append(klines, models.KLine{
			StockCode: code,
			Interval:  interval,
			Time:      stamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})
	}

	// Ensure chronological order so the dataset is consistent.
	sort.Slice(klines, func(i, j int) bool { return klines[i].Time.Before(klines[j].Time) })
	return s.SaveKLines(klines)
}
