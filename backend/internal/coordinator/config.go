package coordinator

import (
	"github.com/christianmz565/microphoto/pkg/env"
)

type Config struct {
	Port           string
	MetricsPort    int
	RedisAddr      string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioSSL       bool
	MaxUploadSize  int64
}

// NewConfig creates a new Config populated from environment variables with default fallbacks.
func NewConfig() *Config {
	return &Config{
		Port:           env.String("PORT", "8080"),
		MetricsPort:    env.Int("METRICS_PORT", 9090),
		RedisAddr:      env.String("REDIS_URL", "localhost:6379"),
		MinioEndpoint:  env.String("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: env.String("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: env.String("MINIO_SECRET_KEY", "minioadmin"),
		MinioSSL:       env.Bool("MINIO_SSL", false),
		MaxUploadSize:  env.Int64("MAX_UPLOAD_SIZE", 50<<20),
	}
}
