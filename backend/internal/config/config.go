package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppID   int
	AppHash string
	Port    string
	CORS    string
}

func Load() *Config {
	_ = godotenv.Load()

	appID, err := strconv.Atoi(os.Getenv("TG_APP_ID"))
	if err != nil {
		log.Fatal("TG_APP_ID must be a valid integer")
	}

	return &Config{
		AppID:   appID,
		AppHash: mustEnv("TG_APP_HASH"),
		Port:    getEnv("PORT", "8080"),
		CORS:    getEnv("CORS_ORIGIN", "*"),
	}
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s is required", k)
	}
	return v
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
