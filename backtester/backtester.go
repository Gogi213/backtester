package backtester

import (
	"fmt"
	"sync"
	"time"
)

// Trade represents a single trade
type Trade struct {
	ID         string    `json:"id"`
	Price      float64   `json:"price"`
	Qty        float64   `json:"qty"`
	Time       time.Time `json:"time"`
	IsBuy      bool      `json:"is_buy"`
	Commission float64   `json:"commission"`
}

// Position represents a current position
type Position struct {
	Symbol        string    `json:"symbol"`
	Qty           float64   `json:"qty"`
	AvgEntryPrice float64   `json:"avg_entry_price"`
	OpenTime      time.Time `json:"open_time"`
}

// Portfolio represents a trading portfolio
type Portfolio struct {
	Cash      float64             `json:"cash"`
	Positions map[string]Position `json:"positions"`
	Equity    float64             `json:"equity"`
}

// Order represents a trading order
type Order struct {
	Symbol string    `json:"symbol"`
	Qty    float64   `json:"qty"`
	Price  float64   `json:"price"`
	IsBuy  bool      `json:"is_buy"`
	Time   time.Time `json:"time"`
}

// CommissionCalculator handles commission calculations
type CommissionCalculator struct {
	rate float64 // Commission rate (0.0005 = 0.05%)
}

// NewCommissionCalculator creates a new commission calculator
func NewCommissionCalculator(rate float64) *CommissionCalculator {
	return &CommissionCalculator{
		rate: rate,
	}
}

// CalculateCommission calculates commission for a trade
func (cc *CommissionCalculator) CalculateCommission(price, qty float64) float64 {
	return price * qty * cc.rate
}

// CalculateTotalCommission calculates total commission for entry and exit
func (cc *CommissionCalculator) CalculateTotalCommission(entryPrice, exitPrice, qty float64) float64 {
	entryCommission := cc.CalculateCommission(entryPrice, qty)
	exitCommission := cc.CalculateCommission(exitPrice, qty)
	return entryCommission + exitCommission
}

// TradeExecutor handles trade execution
type TradeExecutor struct {
	commissionCalculator *CommissionCalculator
}

// NewTradeExecutor creates a new trade executor
func NewTradeExecutor(commissionRate float64) *TradeExecutor {
	return &TradeExecutor{
		commissionCalculator: NewCommissionCalculator(commissionRate),
	}
}

// ExecuteTrade executes a trade and returns the executed trade with commission
func (te *TradeExecutor) ExecuteTrade(symbol string, price, qty float64, isBuy bool, timestamp time.Time) *Trade {
	commission := te.commissionCalculator.CalculateCommission(price, qty)

	return &Trade{
		ID:         generateTradeID(),
		Price:      price,
		Qty:        qty,
		Time:       timestamp,
		IsBuy:      isBuy,
		Commission: commission,
	}
}

var (
	idMutex sync.Mutex
	lastID  int64
)

// generateTradeID generates a unique trade ID
func generateTradeID() string {
	idMutex.Lock()
	defer idMutex.Unlock()

	now := time.Now()
	currentID := now.UnixNano()

	// Ensure ID is unique by incrementing if needed
	if currentID <= lastID {
		currentID = lastID + 1
	}

	lastID = currentID

	return fmt.Sprintf("trade_%d", currentID)
}

// PortfolioManager handles portfolio management
type PortfolioManager struct {
	portfolio            *Portfolio
	commissionCalculator *CommissionCalculator
}

// NewPortfolioManager creates a new portfolio manager
func NewPortfolioManager(initialCash float64, commissionRate float64) *PortfolioManager {
	return &PortfolioManager{
		portfolio: &Portfolio{
			Cash:      initialCash,
			Positions: make(map[string]Position),
			Equity:    initialCash,
		},
		commissionCalculator: NewCommissionCalculator(commissionRate),
	}
}

// GetPortfolio returns the current portfolio
func (pm *PortfolioManager) GetPortfolio() *Portfolio {
	return pm.portfolio
}

// UpdateEquity updates portfolio equity based on current market prices
func (pm *PortfolioManager) UpdateEquity(currentPrices map[string]float64) {
	equity := pm.portfolio.Cash

	// Add value of all positions
	for symbol, position := range pm.portfolio.Positions {
		currentPrice, exists := currentPrices[symbol]
		if exists {
			positionValue := position.Qty * currentPrice
			equity += positionValue
		}
	}

	pm.portfolio.Equity = equity
}

// ExecuteOrder executes an order and updates the portfolio
func (pm *PortfolioManager) ExecuteOrder(order *Order) (*Trade, error) {
	// Calculate commission
	commission := pm.commissionCalculator.CalculateCommission(order.Price, order.Qty)

	// Check if we have enough cash for buy order
	if order.IsBuy {
		cost := order.Price*order.Qty + commission
		if pm.portfolio.Cash < cost {
			return nil, &InsufficientFundsError{Available: pm.portfolio.Cash, Required: cost}
		}
	}

	// Execute trade
	trade := &Trade{
		ID:         generateTradeID(),
		Price:      order.Price,
		Qty:        order.Qty,
		Time:       order.Time,
		IsBuy:      order.IsBuy,
		Commission: commission,
	}

	// Update portfolio
	if order.IsBuy {
		pm.portfolio.Cash -= order.Price*order.Qty + commission
		pm.updatePosition(order.Symbol, order.Qty, order.Price, order.Time, true)
	} else {
		pm.portfolio.Cash += order.Price*order.Qty - commission
		pm.updatePosition(order.Symbol, order.Qty, order.Price, order.Time, false)
	}

	return trade, nil
}

// updatePosition updates a position in the portfolio
func (pm *PortfolioManager) updatePosition(symbol string, qty, price float64, timestamp time.Time, isBuy bool) {
	position, exists := pm.portfolio.Positions[symbol]

	if !exists {
		// Create new position
		if isBuy {
			pm.portfolio.Positions[symbol] = Position{
				Symbol:        symbol,
				Qty:           qty,
				AvgEntryPrice: price,
				OpenTime:      timestamp,
			}
		} else {
			// Short position
			pm.portfolio.Positions[symbol] = Position{
				Symbol:        symbol,
				Qty:           -qty,
				AvgEntryPrice: price,
				OpenTime:      timestamp,
			}
		}
	} else {
		// Update existing position
		if isBuy {
			// Adding to long position or closing short position
			newQty := position.Qty + qty
			if newQty > 0 {
				// Still long
				newAvgPrice := (position.AvgEntryPrice*position.Qty + price*qty) / newQty
				position.Qty = newQty
				position.AvgEntryPrice = newAvgPrice
			} else if newQty == 0 {
				// Position closed
				delete(pm.portfolio.Positions, symbol)
				return
			} else {
				// Now short
				position.Qty = newQty
				position.AvgEntryPrice = price
			}
		} else {
			// Adding to short position or closing long position
			newQty := position.Qty - qty
			if newQty < 0 {
				// Still short
				newAvgPrice := (position.AvgEntryPrice*-position.Qty + price*qty) / -newQty
				position.Qty = newQty
				position.AvgEntryPrice = newAvgPrice
			} else if newQty == 0 {
				// Position closed
				delete(pm.portfolio.Positions, symbol)
				return
			} else {
				// Now long
				position.Qty = newQty
				position.AvgEntryPrice = price
			}
		}

		if position.Qty != 0 {
			pm.portfolio.Positions[symbol] = position
		} else {
			delete(pm.portfolio.Positions, symbol)
		}
	}
}

// InsufficientFundsError represents an error when there are insufficient funds
type InsufficientFundsError struct {
	Available float64
	Required  float64
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("insufficient funds: available $%.2f, required $%.2f", e.Available, e.Required)
}
