// Package env provides helpers for reading environment variables with fallback defaults.
package env

import (
	"os"
	"strconv"
)

// String retrieves the value of the environment variable named by the key.
// It returns the value, which will be the fallback if the variable is not present.
func String(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

// Int retrieves the value of the environment variable named by the key as an integer.
// It returns the value, which will be the fallback if the variable is not present or cannot be parsed.
func Int(key string, fallback int) int {
	valueStr := String(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return fallback
}

// Int64 retrieves the value of the environment variable named by the key as a 64-bit integer.
// It returns the value, which will be the fallback if the variable is not present or cannot be parsed.
func Int64(key string, fallback int64) int64 {
	valueStr := String(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}

	return fallback
}

// Bool retrieves the value of the environment variable named by the key as a boolean.
// It returns the value, which will be the fallback if the variable is not present or cannot be parsed.
func Bool(key string, fallback bool) bool {
	valueStr := String(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}

	return fallback
}
