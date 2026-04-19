package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                   string
	DBDriver               string
	DBDSN                  string
	AllowedOrigin          string
	MockUserID             string
	RateLimitRequests      int
	RateLimitWindow        time.Duration
	RateLimitCleanupWindow time.Duration
}

func Load() Config {
	return Config{
		Port:                   getEnv("PLANAHEAD_PORT", "8000"),
		DBDriver:               getEnv("PLANAHEAD_DB_DRIVER", "mysql"),
		DBDSN:                  getEnv("PLANAHEAD_DB_DSN", "root@tcp(127.0.0.1:3306)/degree_tracker?parseTime=true"),
		AllowedOrigin:          getEnv("PLANAHEAD_ALLOWED_ORIGIN", "http://localhost:3000"),
		MockUserID:             getEnv("PLANAHEAD_MOCK_USER_ID", "local-demo-user"),
		RateLimitRequests:      getEnvInt("PLANAHEAD_RATE_LIMIT_REQUESTS", 10),
		RateLimitWindow:        getEnvDuration("PLANAHEAD_RATE_LIMIT_WINDOW", time.Minute),
		RateLimitCleanupWindow: getEnvDuration("PLANAHEAD_RATE_LIMIT_CLEANUP_WINDOW", 5*time.Minute),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
