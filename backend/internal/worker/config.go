// Package worker handles image/video processing tasks.
package worker

import (
	"github.com/christianmz565/microphoto/pkg/env"
)

// Config holds the configuration for the worker service.
type Config struct {
	MetricsPort    int
	RedisAddr      string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioSSL       bool
}

// NewConfig creates a new Config populated from environment variables with default fallbacks.
func NewConfig() *Config {
	return &Config{
		MetricsPort:    env.Int("METRICS_PORT", 9091),
		RedisAddr:      env.String("REDIS_URL", "localhost:6379"),
		MinioEndpoint:  env.String("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: env.String("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: env.String("MINIO_SECRET_KEY", "minioadmin"),
		MinioSSL:       env.Bool("MINIO_SSL", false),
	}
}
