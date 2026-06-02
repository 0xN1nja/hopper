package config

import "os"

type Config struct {
	Port     string
	DataPath string
	Username string
	Password string
	Dev      bool
}

func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "8080"),
		DataPath: getEnv("DATA_PATH", "/data"),
		Username: getEnv("USERNAME", "admin"),
		Password: getEnv("PASSWORD", ""),
		Dev:      getEnv("DEV", "") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
