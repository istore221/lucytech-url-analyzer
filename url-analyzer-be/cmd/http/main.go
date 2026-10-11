package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/config"
	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/handler"
	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/scraper"
	"github.com/istore221/lucytech-url-analyzer/internal/analyzer"
)

func main() {
	cfg := config.Load()

	engine := scraper.NewService(analyzer.New(
		analyzer.WithWorkers(cfg.AnalyzerWorkerPoolSize),
	))

	scrapeHandler := handler.NewScrapeHandler(engine,
		cfg.ScrapeTimeout,
		cfg.MaxHTTPBatchConcurrency,
		cfg.MaxBatchSize)

	router := gin.Default()

	router.GET("/health", scrapeHandler.Health)
	router.POST("/api/v1/scrape", scrapeHandler.Scrape)

	log.Printf("Server starting on port %s with scrape timeout of %s...", cfg.Port, cfg.ScrapeTimeout)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
