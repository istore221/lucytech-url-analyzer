package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		cfg.MaxHTTPBatchSize)

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/health", scrapeHandler.Health)
	router.POST("/api/v1/scrape", scrapeHandler.Scrape)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on port %s with scrape timeout of %s...", cfg.Port, cfg.ScrapeTimeout)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Give active requests 10 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
