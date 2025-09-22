package backtester

import (
	"hft-backtester/strategies"
	"time"
)

// BacktestResult represents the results of a backtest
type BacktestResult struct {
	Trades      []*Trade      `json:"trades"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	FinalEquity float64       `json:"final_equity"`
	EquityCurve []EquityPoint `json:"equity_curve"`
	PriceData   []ChartPoint  `json:"price_data,omitempty"` // Added for visualization
}

// EquityPoint represents a point in the equity curve
type EquityPoint struct {
	Time   int64   `json:"time"`
	Equity float64 `json:"equity"`
}

// ChartPoint represents a data point for charting
type ChartPoint struct {
	Time  int64   `json:"time"`
	Price float64 `json:"price"`
}

// BacktestEngine represents the backtesting engine
type BacktestEngine struct {
	portfolioManager *PortfolioManager
	tradeExecutor    *TradeExecutor
	positionSize     float64 // Position size in USD
}

// NewBacktestEngine creates a new backtesting engine
func NewBacktestEngine(initialCash float64, commissionRate float64, positionSize float64) *BacktestEngine {
	return &BacktestEngine{
		portfolioManager: NewPortfolioManager(initialCash, commissionRate),
		tradeExecutor:    NewTradeExecutor(commissionRate),
		positionSize:     positionSize,
	}
}

// Run executes a backtest with a given strategy
func (be *BacktestEngine) Run(data []ChartPoint, strategy interface{}) *BacktestResult {
	result := &BacktestResult{
		Trades:      make([]*Trade, 0),
		StartTime:   time.Now(),
		EquityCurve: make([]EquityPoint, 0),
	}

	// Initialize Bollinger Bands strategy if provided
	var bbStrategy *strategies.BollingerBandsStrategy
	if strategy != nil {
		if bb, ok := strategy.(*strategies.BollingerBandsStrategy); ok {
			bbStrategy = bb
		}
	}

	// Process each data point
	for _, point := range data {
		// Update equity curve
		be.portfolioManager.UpdateEquity(map[string]float64{"BTCUSDT": point.Price})
		result.EquityCurve = append(result.EquityCurve, EquityPoint{
			Time:   point.Time,
			Equity: be.portfolioManager.GetPortfolio().Equity,
		})

		var order *Order
		if bbStrategy != nil {
			// Use strategy to determine action
			signal := bbStrategy.GetSignal(point.Price)

			// Check current position
			currentPosition := 0.0
			if pos, exists := be.portfolioManager.GetPortfolio().Positions["BTCUSDT"]; exists {
				currentPosition = pos.Qty
			}

			// Create order based on signal
			switch signal.Action {
			case "BUY":
				// Only buy if we don't have a position
				if currentPosition <= 0 && be.portfolioManager.GetPortfolio().Cash >= be.positionSize {
					qty := be.positionSize / point.Price
					order = &Order{
						Symbol: "BTCUSDT",
						Qty:    qty,
						Price:  point.Price,
						IsBuy:  true,
						Time:   time.Unix(point.Time/1000, 0),
					}
				}
			case "SELL":
				// Only sell if we have a long position
				if currentPosition > 0 {
					qty := currentPosition // Close entire position
					order = &Order{
						Symbol: "BTCUSDT",
						Qty:    qty,
						Price:  point.Price,
						IsBuy:  false,
						Time:   time.Unix(point.Time/1000, 0),
					}
				}
			case "EXIT_LONG":
				// Exit long position if we have one
				if currentPosition > 0 {
					order = &Order{
						Symbol: "BTCUSDT",
						Qty:    currentPosition, // Close entire position
						Price:  point.Price,
						IsBuy:  false,
						Time:   time.Unix(point.Time/1000, 0),
					}
				}
			case "EXIT_SHORT":
				// Exit short position if we have one
				if currentPosition < 0 {
					order = &Order{
						Symbol: "BTCUSDT",
						Qty:    -currentPosition, // Close entire position
						Price:  point.Price,
						IsBuy:  true,
						Time:   time.Unix(point.Time/1000, 0),
					}
				}
			}
		} else {
			// Default behavior - buy all points (for backward compatibility)
			qty := be.positionSize / point.Price
			order = &Order{
				Symbol: "BTCUSDT",
				Qty:    qty,
				Price:  point.Price,
				IsBuy:  true,
				Time:   time.Unix(point.Time/1000, 0),
			}
		}

		// Execute the order if one was created
		if order != nil {
			trade, err := be.portfolioManager.ExecuteOrder(order)
			if err != nil {
				// Log error but continue
				continue
			}
			result.Trades = append(result.Trades, trade)
		}
	}

	result.EndTime = time.Now()
	result.FinalEquity = be.portfolioManager.GetPortfolio().Equity

	return result
}
