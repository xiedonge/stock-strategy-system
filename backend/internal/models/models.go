package models

import "time"

// Stock represents a tradable A-share equity.
type Stock struct {
	ID        uint      `gorm:"primaryKey"`
	Code      string    `gorm:"uniqueIndex;size:16"`
	Name      string    `gorm:"size:64"`
	Exchange  string    `gorm:"size:16"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// KLine holds OHLCV data for a given interval.
type KLine struct {
	ID        uint      `gorm:"primaryKey"`
	StockCode string    `gorm:"index;size:16"`
	Interval  string    `gorm:"index;size:8"` // 1d or 30m
	Time      time.Time `gorm:"index"`
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CreatedAt time.Time
}

// Strategy defines a screening rule and its serialized parameters.
type Strategy struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"size:64"`
	Description string    `gorm:"size:256"`
	Type        string    `gorm:"size:32"` // e.g. ma_crossover
	ParamsJSON  string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Backtest stores summary results for a strategy on a single stock.
type Backtest struct {
	ID             uint      `gorm:"primaryKey"`
	StrategyID     uint      `gorm:"index"`
	StockCode      string    `gorm:"size:16;index"`
	Start          time.Time
	End            time.Time
	InitialCapital float64
	FinalCapital   float64
	ReturnPct      float64
	CreatedAt      time.Time
}

// BacktestPoint is a single point on the equity curve.
type BacktestPoint struct {
	ID         uint      `gorm:"primaryKey"`
	BacktestID uint      `gorm:"index"`
	Time       time.Time `gorm:"index"`
	Equity     float64
}
