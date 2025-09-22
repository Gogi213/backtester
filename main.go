package main

import (
	"hft-backtester/handlers"
	"hft-backtester/templates"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
)

func main() {
	app := fiber.New()

	// Add gzip compression for faster data transfer
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Serve static files from templates directory
	app.Static("/templates", "./templates")

	app.Get("/", func(c *fiber.Ctx) error {
		html := templates.GetHTMLTemplate()
		return c.Type("html").SendString(html)
	})

	app.Get("/api/trades", handlers.GetTradesHandler)
	app.Get("/api/hours", handlers.GetHoursHandler)
	app.Post("/api/backtest", handlers.RunBacktestHandler)
	app.Get("/health", handlers.HealthHandler)

	log.Printf("Server starting on port 8080...")
	log.Fatal(app.Listen(":8080"))
}
