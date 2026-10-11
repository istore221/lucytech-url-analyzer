package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCleanTestEnv(t *testing.T) {
	t.Helper()

	viper.Reset()

	tempDir := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		viper.Reset()
		_ = os.Chdir(origWD)
	})
}

func TestLoad_Defaults(t *testing.T) {
	setupCleanTestEnv(t)

	cfg := Load()

	assert.NotNil(t, cfg)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, 20*time.Second, cfg.ScrapeTimeout)
	assert.Equal(t, 10, cfg.MaxHTTPBatchConcurrency)
	assert.Equal(t, 1, cfg.MaxHTTPBatchSize)
	assert.Equal(t, 10, cfg.AnalyzerWorkerPoolSize)
	assert.Equal(t, "DEBUG", cfg.LogLevel)
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	setupCleanTestEnv(t)

	t.Setenv("PORT", "9090")
	t.Setenv("SCRAPE_TIMEOUT", "45s")
	t.Setenv("MAX_HTTP_BATCH_CONCURRENCY", "20")
	t.Setenv("MAX_HTTP_BATCH_SIZE", "5")
	t.Setenv("ANALYZER_WORKER_POOL_SIZE", "15")
	t.Setenv("HTTP_TIMEOUT", "10s")
	t.Setenv("LOG_LEVEL", "INFO")

	cfg := Load()

	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, 45*time.Second, cfg.ScrapeTimeout)
	assert.Equal(t, 20, cfg.MaxHTTPBatchConcurrency)
	assert.Equal(t, 5, cfg.MaxHTTPBatchSize)
	assert.Equal(t, 15, cfg.AnalyzerWorkerPoolSize)
	assert.Equal(t, 10*time.Second, cfg.HTTPTimeout)
	assert.Equal(t, "INFO", cfg.LogLevel)
}

func TestLoad_ConfigFile(t *testing.T) {
	setupCleanTestEnv(t)

	envContent := []byte(`
PORT=3000
SCRAPE_TIMEOUT=15s
MAX_HTTP_BATCH_CONCURRENCY=8
MAX_HTTP_BATCH_SIZE=3
ANALYZER_WORKER_POOL_SIZE=5
HTTP_TIMEOUT=12s
LOG_LEVEL=WARN
`)
	err := os.WriteFile(".env", envContent, 0644)
	require.NoError(t, err)

	cfg := Load()

	assert.Equal(t, "3000", cfg.Port)
	assert.Equal(t, 15*time.Second, cfg.ScrapeTimeout)
	assert.Equal(t, 8, cfg.MaxHTTPBatchConcurrency)
	assert.Equal(t, 3, cfg.MaxHTTPBatchSize)
	assert.Equal(t, 5, cfg.AnalyzerWorkerPoolSize)
	assert.Equal(t, 12*time.Second, cfg.HTTPTimeout)
	assert.Equal(t, "WARN", cfg.LogLevel)
}

func TestLoad_EnvVarOverridesConfigFile(t *testing.T) {
	setupCleanTestEnv(t)

	envContent := []byte(`
PORT=3000
LOG_LEVEL=WARN
`)
	err := os.WriteFile(".env", envContent, 0644)
	require.NoError(t, err)

	t.Setenv("PORT", "5000")

	cfg := Load()

	assert.Equal(t, "5000", cfg.Port)     // Overridden by env var
	assert.Equal(t, "WARN", cfg.LogLevel) // Read from .env file
}
