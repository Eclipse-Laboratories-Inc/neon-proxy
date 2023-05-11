package config

import (
	"github.com/test-go/testify/assert"
	"testing"
)

func TestCreateConfigFromEnv(t *testing.T) {
	t.Run("success load config with default values", func(t *testing.T) {
		const defaultMemCapacity = 4
		cfg, err := CreateConfigFromEnv("test", ".test.env")
		assert.NoError(t, err)
		assert.Equal(t, "http://localhost:8899", cfg.SolanaEndpoint)
		assert.Equal(t, "http://localhost:9000", cfg.PpSolanaEndpoint)
		assert.Equal(t, defaultMemCapacity, cfg.MempoolCapacity)
	})
	t.Run("fail load config: MempoolCapacity exceeds maximum allowed value", func(t *testing.T) {
		_, err := CreateConfigFromEnv("test", ".test2.env")
		assert.Error(t, err)
	})
}
