package handlers

import (
	"encoding/csv"
	"hft-backtester/backtester"
	"hft-backtester/strategies"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Trade struct {
	ID           string `json:"id"`
	Price        string `json:"price"`
	Qty          string `json:"qty"`
	QuoteQty     string `json:"quote_qty"`
	Time         string `json:"time"`
	IsBuyerMaker string `json:"is_buyer_maker"`
}

type ChartPoint struct {
	Time  int64   `json:"time"`
	Price float64 `json:"price"`
}

type HourInfo struct {
	Hour  string `json:"hour"`
	Count int    `json:"count"`
}

// BacktestRequest represents the parameters for a backtest
type BacktestRequest struct {
	Strategy       string                 `json:"strategy"`
	InitialCash    float64                `json:"initial_cash"`
	PositionSize   float64                `json:"position_size"`
	Commission     float64                `json:"commission"`
	Hour           string                 `json:"hour"`
	StartTime      string                 `json:"start_time"`
	EndTime        string                 `json:"end_time"`
	StrategyParams map[string]interface{} `json:"strategy_params"`
}

func LoadTradesByHour(hour string) ([]ChartPoint, error) {
	return LoadTradesByHourWithLimit(hour, 10000) // Limit to 10,000 points by default
}

func LoadTradesByHourWithLimit(hour string, limit int) ([]ChartPoint, error) {
	file, err := os.Open("upload/trades/STBLUSDT-trades-2025-09-20.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var points []ChartPoint
	reader := csv.NewReader(file)

	// Read and skip header
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	// For sampling data when we have more points than limit
	sampleRate := 1
	count := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(record) >= 5 {
			timestamp, _ := strconv.ParseInt(record[4], 10, 64)
			recordTime := time.Unix(timestamp/1000, 0)
			recordHour := recordTime.Format("15")

			if recordHour == hour {
				count++
				// Sample data if we have more points than limit
				if limit > 0 && count > limit {
					if sampleRate == 1 {
						// Calculate sample rate based on actual count and limit
						// We need to re-read the file to get actual count, so we'll use a simple approach
						// For now, we'll just increase sample rate and continue
						sampleRate = count / limit
						if sampleRate < 2 {
							sampleRate = 2
						}
					}

					// Only add point if it matches sample rate
					if count%sampleRate != 0 {
						continue
					}
				}

				price, _ := strconv.ParseFloat(record[1], 64)
				points = append(points, ChartPoint{
					Time:  timestamp,
					Price: price,
				})
			}
		}
	}

	return points, nil
}

func GetAvailableHours() ([]HourInfo, error) {
	file, err := os.Open("upload/trades/STBLUSDT-trades-2025-09-20.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hourCounts := make(map[string]int)
	reader := csv.NewReader(file)

	// Read and skip header
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(record) >= 5 {
			timestamp, _ := strconv.ParseInt(record[4], 10, 64)
			recordTime := time.Unix(timestamp/1000, 0)
			hour := recordTime.Format("15")
			hourCounts[hour]++
		}
	}

	var hours []HourInfo
	for hour, count := range hourCounts {
		hours = append(hours, HourInfo{
			Hour:  hour,
			Count: count,
		})
	}

	return hours, nil
}

func LoadTrades(limit int) ([]ChartPoint, error) {
	file, err := os.Open("upload/trades/STBLUSDT-trades-2025-09-20.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var points []ChartPoint
	reader := csv.NewReader(file)

	// Read and skip header
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	count := 0
	for limit == 0 || len(points) < limit {
		record, err := reader.Read()
		if err == io.EOF {
			// EOF - stop reading
			break
		}
		if err != nil {
			return nil, err
		}

		// record format: [id, price, qty, quote_qty, time, is_buyer_maker]
		if len(record) >= 5 {
			price, _ := strconv.ParseFloat(record[1], 64)
			time, _ := strconv.ParseInt(record[4], 10, 64)

			points = append(points, ChartPoint{
				Time:  time,
				Price: price,
			})
		}

		count++
	}

	return points, nil
}

// GetTradesHandler handles requests for trades data
func GetTradesHandler(c *fiber.Ctx) error {
	hour := c.Query("hour")
	if hour != "" {
		trades, err := LoadTradesByHour(hour)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(trades)
	}

	trades, err := LoadTrades(1000)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(trades)
}

// GetHoursHandler handles requests for available hours
func GetHoursHandler(c *fiber.Ctx) error {
	hours, err := GetAvailableHours()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(hours)
}

// RunBacktestHandler handles backtest requests
func RunBacktestHandler(c *fiber.Ctx) error {
	var req BacktestRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request format"})
	}

	// Load data for backtesting
	var trades []ChartPoint
	var err error
	if req.Hour != "" {
		trades, err = LoadTradesByHour(req.Hour)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	} else {
		// Load default 1000 points if no hour specified
		trades, err = LoadTrades(1000)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}
	// Create backtest engine
	initialCash := req.InitialCash
	if initialCash <= 0 {
		initialCash = 100.0 // Default value
	}

	commission := req.Commission
	if commission <= 0 {
		commission = 0.05 // Default value (0.05%)
	}

	positionSize := req.PositionSize
	if positionSize <= 0 {
		positionSize = 50.0 // Default value ($50)
	}

	engine := backtester.NewBacktestEngine(initialCash, commission/100.0, positionSize) // Convert percentage to decimal

	// Convert ChartPoint to backtester.ChartPoint
	backtesterTrades := make([]backtester.ChartPoint, len(trades))
	for i, trade := range trades {
		backtesterTrades[i] = backtester.ChartPoint{
			Time:  trade.Time,
			Price: trade.Price,
		}
	}

	// Create strategy if specified
	var strategy interface{}
	if req.Strategy == "bollinger" {
		period := 100
		stdDev := 1.0

		if p, ok := req.StrategyParams["period"].(float64); ok {
			period = int(p)
		}
		if s, ok := req.StrategyParams["stdDev"].(float64); ok {
			stdDev = s
		}

		strategy = strategies.NewBollingerBandsStrategy(period, stdDev)
	}

	result := engine.Run(backtesterTrades, strategy)

	// Add price data for visualization
	result.PriceData = backtesterTrades

	return c.JSON(result)
}

// HealthHandler handles health check requests
func HealthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "hft-backtester",
	})
}
