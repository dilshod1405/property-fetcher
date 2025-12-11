package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PFAPIUrl    string
	PFAPIKey    string
	PFAPISecret string
	PostgresDSN string
}

var AppConfig *Config

func LoadConfig() {
	_ = godotenv.Load()

	AppConfig = &Config{
		PFAPIUrl:    getEnv("PF_API_URL", ""),
		PFAPIKey:    getEnv("PF_API_KEY", ""),
		PFAPISecret: getEnv("PF_API_SECRET", ""),
		PostgresDSN: getEnv("POSTGRES_DSN", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
