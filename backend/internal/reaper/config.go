// Package reaper handles timeout detection and task rescheduling.
package reaper

import (
	"github.com/christianmz565/microphoto/pkg/env"
)

// Config holds the configuration for the reaper service.
type Config struct {
	RedisAddr            string
	MetricsPort          int
	GlobalTimeoutSeconds int64
	IntervalSeconds      int
}

// NewConfig creates a new Config populated from environment variables with default fallbacks.
func NewConfig() *Config {
	return &Config{
		RedisAddr:            env.String("REDIS_URL", "localhost:6379"),
		MetricsPort:          env.Int("METRICS_PORT", 9092),
		GlobalTimeoutSeconds: env.Int64("GLOBAL_TIMEOUT_SECONDS", 30),
		IntervalSeconds:      env.Int("REAPER_INTERVAL_SECONDS", 5),
	}
}
