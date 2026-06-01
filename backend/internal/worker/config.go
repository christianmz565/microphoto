package worker

import (
	"os"
	"strconv"
)

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
		MetricsPort:    getEnvInt("METRICS_PORT", 9091),
		RedisAddr:      getEnv("REDIS_URL", "localhost:6379"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioSSL:       getEnvBool("MINIO_SSL", false),
	}
}

// getEnv retrieves the value of the environment variable named by the key.
// It returns the value, which will be the fallback if the variable is not present.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvInt retrieves the value of the environment variable named by the key as an integer.
// It returns the value, which will be the fallback if the variable is not present or cannot be parsed.
func getEnvInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}

// getEnvBool retrieves the value of the environment variable named by the key as a boolean.
// It returns the value, which will be the fallback if the variable is not present or cannot be parsed.
func getEnvBool(key string, fallback bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return fallback
}
