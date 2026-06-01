package coordinator

import (
	"os"
	"strconv"
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

func NewConfig() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		MetricsPort:    getEnvInt("METRICS_PORT", 9090),
		RedisAddr:      getEnv("REDIS_URL", "localhost:6379"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioSSL:       getEnvBool("MINIO_SSL", false),
		MaxUploadSize:  getEnvInt64("MAX_UPLOAD_SIZE", 50<<20),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return fallback
}
