package services

import (
	"fmt"

	"github.com/xiedonge/stock-strategy-system/backend/internal/models"
	"github.com/xiedonge/stock-strategy-system/backend/internal/strategy"
	"gorm.io/gorm"
)

// ScreeningResult represents a stock that satisfies a strategy.
type ScreeningResult struct {
	Stock   models.Stock        `json:"stock"`
	Reason  string              `json:"reason"`
	Metrics map[string]float64  `json:"metrics"`
}

// BacktestResult packages the equity curve and summary.
type BacktestResult struct {
	Summary models.Backtest     `json:"summary"`
	Points  []strategy.EquityPoint `json:"points"`
	Trades  []strategy.Trade       `json:"trades"`
}

// AnalysisService performs screening and backtesting.
type AnalysisService struct {
	db       *gorm.DB
	stocks   *StockService
	strategies *StrategyService
}

// NewAnalysisService creates an AnalysisService.
func NewAnalysisService(db *gorm.DB, stocks *StockService, strategies *StrategyService) *AnalysisService {
	return &AnalysisService{db: db, stocks: stocks, strategies: strategies}
}

// Screen runs a strategy across all stocks and returns the matches.
func (a *AnalysisService) Screen(strategyID uint) ([]ScreeningResult, error) {
	strategyModel, err := a.strategies.Get(strategyID)
	if err != nil {
		return nil, err
	}

	stocks, err := a.stocks.ListStocks()
	if err != nil {
		return nil, err
	}

	params := strategy.ParseMACrossoverParams(strategyModel.ParamsJSON)
	var results []ScreeningResult
	for _, stock := range stocks {
		klines, err := a.stocks.GetKLines(stock.Code, "1d", 200)
		if err != nil {
			return nil, err
		}
		if strategy.ShouldSelect(klines, params) {
			results = append(results, ScreeningResult{
				Stock:  stock,
				Reason: fmt.Sprintf("MA%d/MA%d 上穿", params.ShortWindow, params.LongWindow),
				Metrics: map[string]float64{
					"short_window": float64(params.ShortWindow),
					"long_window":  float64(params.LongWindow),
				},
			})
		}
	}

	return results, nil
}

// RunBacktest performs a backtest for a given stock and strategy.
func (a *AnalysisService) RunBacktest(strategyID uint, code string, initial float64) (*BacktestResult, error) {
	strategyModel, err := a.strategies.Get(strategyID)
	if err != nil {
		return nil, err
	}

	if initial <= 0 {
		initial = 100000
	}

	klines, err := a.stocks.GetKLines(code, "1d", 1000)
	if err != nil {
		return nil, err
	}
	params := strategy.ParseMACrossoverParams(strategyModel.ParamsJSON)
	final, points, trades := strategy.Backtest(klines, params, initial)
	if len(klines) == 0 {
		return nil, fmt.Errorf("no kline data for %s", code)
	}

	summary := models.Backtest{
		StrategyID:     strategyID,
		StockCode:      code,
		Start:          klines[0].Time,
		End:            klines[len(klines)-1].Time,
		InitialCapital: initial,
		FinalCapital:   final,
		ReturnPct:      (final - initial) / initial * 100,
	}

	if err := a.db.Create(&summary).Error; err != nil {
		return nil, err
	}

	for _, point := range points {
		a.db.Create(&models.BacktestPoint{BacktestID: summary.ID, Time: point.Time, Equity: point.Equity})
	}

	return &BacktestResult{Summary: summary, Points: points, Trades: trades}, nil
}
