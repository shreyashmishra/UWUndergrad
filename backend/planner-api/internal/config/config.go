package config

import "os"

type Config struct {
	Port          string
	DBDriver      string
	DBDSN         string
	AllowedOrigin string
	MockUserID    string
}

func Load() Config {
	return Config{
		Port:          getEnv("PLANAHEAD_PORT", "8080"),
		DBDriver:      getEnv("PLANAHEAD_DB_DRIVER", "mysql"),
		DBDSN:         getEnv("PLANAHEAD_DB_DSN", "root@tcp(127.0.0.1:3306)/degree_tracker?parseTime=true"),
		AllowedOrigin: getEnv("PLANAHEAD_ALLOWED_ORIGIN", "http://localhost:3000"),
		MockUserID:    getEnv("PLANAHEAD_MOCK_USER_ID", "local-demo-user"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
