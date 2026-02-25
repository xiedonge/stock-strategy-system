package handlers

import "github.com/xiedonge/stock-strategy-system/backend/internal/models"

// StrategyRequest defines the payload to create/update a strategy.
type StrategyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	ParamsJSON  string `json:"params_json"`
}

func (s StrategyRequest) toModel() *models.Strategy {
	return &models.Strategy{
		Name:        s.Name,
		Description: s.Description,
		Type:        s.Type,
		ParamsJSON:  s.ParamsJSON,
	}
}

func (s StrategyRequest) apply(strategy *models.Strategy) {
	if strategy == nil {
		return
	}
	strategy.Name = s.Name
	strategy.Description = s.Description
	strategy.Type = s.Type
	strategy.ParamsJSON = s.ParamsJSON
}

// ScreenRequest defines the payload for screening.
type ScreenRequest struct {
	StrategyID uint `json:"strategy_id"`
}

// BacktestRequest defines the payload for running a backtest.
type BacktestRequest struct {
	StrategyID     uint    `json:"strategy_id"`
	StockCode      string  `json:"stock_code"`
	InitialCapital float64 `json:"initial_capital"`
}

// AkshareSyncRequest defines the payload for AkShare data sync.
type AkshareSyncRequest struct {
	Symbols   []string `json:"symbols"`
	Mode      string   `json:"mode"`
	StartDate string   `json:"start_date"`
	EndDate   string   `json:"end_date"`
	MinStart  string   `json:"min_start"`
	MinEnd    string   `json:"min_end"`
	Period    string   `json:"period"`
	Limit     int      `json:"limit"`
}
