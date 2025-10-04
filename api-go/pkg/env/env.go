package env

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/joho/godotenv"
)

func GetString(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	return value
}

func GetInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}

	return intValue
}

func GetByteSize(key string, defaultValue int64) int64 {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	intValue, err := ParseByteSize(value)
	if err != nil {
		panic(err)
	}

	return intValue
}

func GetBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		panic(err)
	}

	return boolValue
}

func GetDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		panic(fmt.Errorf("invalid duration value for %s: %s", key, value))
	}
	return defaultValue
}

func GetSlice(value string) []string {
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func ParseByteSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, errors.New("empty string")
	}

	lastChar := rune(s[len(s)-1])
	var numStr, unit string

	if unicode.IsLetter(lastChar) {
		numStr = s[:len(s)-1]
		unit = strings.ToLower(string(lastChar))
	} else {
		numStr = s
		unit = ""
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %v", err)
	}

	if num < 0 {
		return 0, errors.New("negative values are not allowed")
	}

	switch unit {
	case "k":
		return int64(num * 1024), nil
	case "m":
		return int64(num * 1024 * 1024), nil
	case "g":
		return int64(num * 1024 * 1024 * 1024), nil
	case "t":
		return int64(num * 1024 * 1024 * 1024 * 1024), nil
	case "":
		return int64(num), nil
	default:
		return 0, fmt.Errorf("invalid unit: %s (use k, m, g, t)", unit)
	}
}

func LoadConfigFiles(env string) {
	configFiles := []string{
		".env",
	}

	env = strings.ToLower(env)

	switch env {
	case "development", "dev", "debug":
		configFiles = append(configFiles, ".env.dev")
	case "production", "prod", "release":
		configFiles = append(configFiles, ".env.prod")
	case "test":
		configFiles = append(configFiles, ".env.test")
	}

	configFiles = append(configFiles, ".env.local")

	for _, file := range configFiles {
		if fileExists(file) {
			if err := godotenv.Load(file); err != nil {
				panic(fmt.Errorf("failed to load %s: %w", file, err))
			}
		}
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
