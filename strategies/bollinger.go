package strategies

import (
	"time"
)

// BollingerBandsStrategy represents a Bollinger Bands trading strategy
type BollingerBandsStrategy struct {
	period    int
	stdDev    float64
	sma       []float64
	upperBand []float64
	lowerBand []float64
	prices    []float64
}

// NewBollingerBandsStrategy creates a new Bollinger Bands strategy
func NewBollingerBandsStrategy(period int, stdDev float64) *BollingerBandsStrategy {
	return &BollingerBandsStrategy{
		period: period,
		stdDev: stdDev,
	}
}

// CalculateSMA calculates Simple Moving Average
func (b *BollingerBandsStrategy) CalculateSMA(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}

	sum := 0.0
	for _, price := range prices {
		sum += price
	}
	return sum / float64(len(prices))
}

// CalculateStdDev calculates standard deviation
func (b *BollingerBandsStrategy) CalculateStdDev(prices []float64, mean float64) float64 {
	if len(prices) == 0 {
		return 0
	}

	sum := 0.0
	for _, price := range prices {
		diff := price - mean
		sum += diff * diff
	}
	variance := sum / float64(len(prices))
	return sqrt(variance) // Return standard deviation, not variance
}

// sqrt calculates square root using Newton's method
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}

	z := x / 2
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

// ShouldEnterLong checks if we should enter a long position
func (b *BollingerBandsStrategy) ShouldEnterLong(currentPrice float64) bool {
	if len(b.sma) == 0 || len(b.lowerBand) == 0 {
		return false
	}

	lastLowerBand := b.lowerBand[len(b.lowerBand)-1]

	// Enter long when price is below lower band
	return currentPrice < lastLowerBand
}

// ShouldEnterShort checks if we should enter a short position
func (b *BollingerBandsStrategy) ShouldEnterShort(currentPrice float64) bool {
	if len(b.sma) == 0 || len(b.upperBand) == 0 {
		return false
	}

	lastUpperBand := b.upperBand[len(b.upperBand)-1]

	// Enter short when price is above upper band
	return currentPrice > lastUpperBand
}

// ShouldExitLong checks if we should exit a long position
func (b *BollingerBandsStrategy) ShouldExitLong(currentPrice float64) bool {
	if len(b.sma) == 0 {
		return false
	}

	lastSMA := b.sma[len(b.sma)-1]

	// Exit long when price touches or goes above SMA
	return currentPrice >= lastSMA
}

// ShouldExitShort checks if we should exit a short position
func (b *BollingerBandsStrategy) ShouldExitShort(currentPrice float64) bool {
	if len(b.sma) == 0 {
		return false
	}

	lastSMA := b.sma[len(b.sma)-1]

	// Exit short when price touches or goes below SMA
	return currentPrice <= lastSMA
}

// Update updates the strategy with new price data
func (b *BollingerBandsStrategy) Update(price float64) {
	b.prices = append(b.prices, price)

	// Keep only the last 'period' prices
	if len(b.prices) > b.period {
		b.prices = b.prices[1:]
	}

	// Calculate SMA if we have enough data
	if len(b.prices) >= b.period {
		sma := b.CalculateSMA(b.prices)
		b.sma = append(b.sma, sma)

		// Calculate standard deviation
		stdDev := b.CalculateStdDev(b.prices, sma)

		// Calculate bands
		upper := sma + b.stdDev*stdDev
		lower := sma - b.stdDev*stdDev

		b.upperBand = append(b.upperBand, upper)
		b.lowerBand = append(b.lowerBand, lower)

		// Keep only the last values
		if len(b.sma) > b.period {
			b.sma = b.sma[1:]
			b.upperBand = b.upperBand[1:]
			b.lowerBand = b.lowerBand[1:]
		}
	}
}

// GetSignal returns the trading signal based on current price
func (b *BollingerBandsStrategy) GetSignal(currentPrice float64) Signal {
	b.Update(currentPrice)

	if b.ShouldEnterLong(currentPrice) {
		return Signal{Action: "BUY", Price: currentPrice}
	} else if b.ShouldEnterShort(currentPrice) {
		return Signal{Action: "SELL", Price: currentPrice}
	} else if b.ShouldExitLong(currentPrice) {
		return Signal{Action: "EXIT_LONG", Price: currentPrice}
	} else if b.ShouldExitShort(currentPrice) {
		return Signal{Action: "EXIT_SHORT", Price: currentPrice}
	}

	return Signal{Action: "HOLD", Price: currentPrice}
}

// Signal represents a trading signal
type Signal struct {
	Action string
	Price  float64
	Time   time.Time
}
