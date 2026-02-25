package strategy

import (
	"encoding/json"
	"math"
	"sort"
	"time"

	"github.com/xiedonge/stock-strategy-system/backend/internal/models"
)

// MACrossoverParams configures the moving average crossover strategy.
type MACrossoverParams struct {
	ShortWindow int `json:"short_window"`
	LongWindow  int `json:"long_window"`
}

// DefaultMACrossoverParams provides conservative defaults for screening.
func DefaultMACrossoverParams() MACrossoverParams {
	return MACrossoverParams{ShortWindow: 5, LongWindow: 20}
}

// ParseMACrossoverParams parses JSON params, falling back to defaults.
func ParseMACrossoverParams(raw string) MACrossoverParams {
	params := DefaultMACrossoverParams()
	if raw == "" {
		return params
	}

	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		return DefaultMACrossoverParams()
	}

	if params.ShortWindow <= 0 {
		params.ShortWindow = 5
	}
	if params.LongWindow <= 0 {
		params.LongWindow = 20
	}
	if params.ShortWindow >= params.LongWindow {
		params.LongWindow = params.ShortWindow + 5
	}
	return params
}

// ShouldSelect checks whether the most recent data indicates a bullish crossover.
func ShouldSelect(klines []models.KLine, params MACrossoverParams) bool {
	if len(klines) < params.LongWindow+1 {
		return false
	}
	sorted := make([]models.KLine, len(klines))
	copy(sorted, klines)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Time.Before(sorted[j].Time) })

	shortMA := movingAverage(sorted, params.ShortWindow)
	longMA := movingAverage(sorted, params.LongWindow)

	if len(shortMA) < 2 || len(longMA) < 2 {
		return false
	}

	// Cross above: previous short <= long and current short > long.
	prev := shortMA[len(shortMA)-2] - longMA[len(longMA)-2]
	curr := shortMA[len(shortMA)-1] - longMA[len(longMA)-1]
	return prev <= 0 && curr > 0
}

// Backtest runs a simple long-only crossover backtest.
func Backtest(klines []models.KLine, params MACrossoverParams, initial float64) (float64, []EquityPoint, []Trade) {
	if initial <= 0 {
		initial = 100000
	}

	sorted := make([]models.KLine, len(klines))
	copy(sorted, klines)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Time.Before(sorted[j].Time) })

	shortMA := movingAverage(sorted, params.ShortWindow)
	longMA := movingAverage(sorted, params.LongWindow)

	cash := initial
	position := 0.0
	var points []EquityPoint
	var trades []Trade

	for i := range sorted {
		if i < params.LongWindow-1 {
			points = append(points, EquityPoint{Time: sorted[i].Time, Equity: cash})
			continue
		}

		idx := i - (params.LongWindow - 1)
		if idx >= 1 && idx < len(shortMA) && idx < len(longMA) {
			prevDiff := shortMA[idx-1] - longMA[idx-1]
			currDiff := shortMA[idx] - longMA[idx]

			if position == 0 && prevDiff <= 0 && currDiff > 0 {
				// Buy at close.
				price := sorted[i].Close
				shares := math.Floor(cash / price)
				if shares > 0 {
					position = shares
					cash -= shares * price
					trades = append(trades, Trade{Time: sorted[i].Time, Side: "BUY", Price: price, Shares: shares})
				}
			}
			if position > 0 && prevDiff >= 0 && currDiff < 0 {
				// Sell at close.
				price := sorted[i].Close
				cash += position * price
				trades = append(trades, Trade{Time: sorted[i].Time, Side: "SELL", Price: price, Shares: position})
				position = 0
			}
		}

		equity := cash + position*sorted[i].Close
		points = append(points, EquityPoint{Time: sorted[i].Time, Equity: equity})
	}

	final := cash
	if len(sorted) > 0 {
		final += position * sorted[len(sorted)-1].Close
	}

	return final, points, trades
}

// EquityPoint captures one point on the equity curve.
type EquityPoint struct {
	Time   time.Time `json:"time"`
	Equity float64   `json:"equity"`
}

// Trade records a simulated trade decision.
type Trade struct {
	Time   time.Time `json:"time"`
	Side   string    `json:"side"`
	Price  float64   `json:"price"`
	Shares float64   `json:"shares"`
}

func movingAverage(klines []models.KLine, window int) []float64 {
	if window <= 0 || len(klines) < window {
		return nil
	}

	result := make([]float64, 0, len(klines)-window+1)
	var sum float64
	for i := range klines {
		sum += klines[i].Close
		if i >= window {
			sum -= klines[i-window].Close
		}
		if i >= window-1 {
			result = append(result, sum/float64(window))
		}
	}
	return result
}
