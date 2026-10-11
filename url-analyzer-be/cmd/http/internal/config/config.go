package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port                    string        `mapstructure:"PORT"`
	ScrapeTimeout           time.Duration `mapstructure:"SCRAPE_TIMEOUT"`
	MaxHTTPBatchSize        int           `mapstructure:"MAX_HTTP_BATCH_SIZE"`
	MaxHTTPBatchConcurrency int           `mapstructure:"MAX_HTTP_BATCH_CONCURRENCY"`
	AnalyzerWorkerPoolSize  int           `mapstructure:"ANALYZER_WORKER_POOL_SIZE"`
	HTTPTimeout             time.Duration `mapstructure:"HTTP_TIMEOUT"`
	LogLevel                string        `mapstructure:"LOG_LEVEL"`
}

func Load() *Config {

	viper.SetDefault("PORT", "8080")
	viper.SetDefault("SCRAPE_TIMEOUT", "20s")
	viper.SetDefault("MAX_HTTP_BATCH_CONCURRENCY", 10)
	viper.SetDefault("MAX_HTTP_BATCH_SIZE", 1)
	viper.SetDefault("ANALYZER_WORKER_POOL_SIZE", 10)
	viper.SetDefault("HTTPTimeout", "20s")
	viper.SetDefault("LOG_LEVEL", "DEBUG")

	viper.AutomaticEnv()

	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AddConfigPath(".")
	viper.AddConfigPath("./cmd/http")
	viper.AddConfigPath("../..")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Warning: error reading config file: %v", err)
		}
	}

	cfg := &Config{
		Port:                    viper.GetString("PORT"),
		ScrapeTimeout:           viper.GetDuration("SCRAPE_TIMEOUT"),
		MaxHTTPBatchConcurrency: viper.GetInt("MAX_HTTP_BATCH_CONCURRENCY"),
		MaxHTTPBatchSize:        viper.GetInt("MAX_HTTP_BATCH_SIZE"),
		AnalyzerWorkerPoolSize:  viper.GetInt("ANALYZER_WORKER_POOL_SIZE"),
		HTTPTimeout:             viper.GetDuration("HTTP_TIMEOUT"),
		LogLevel:                viper.GetString("LOG_LEVEL"),
	}

	return cfg
}
