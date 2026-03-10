package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL     string `env:"DATABASE_URL" env-required:"true"`
	ServerPort      string `env:"SERVER_PORT" env-default:"8080"`
	WorkerCount     int    `env:"WORKER_COUNT" env-default:"5"`
	HistoricalCapID int    `env:"HISTORICAL_HARD_CAP_ID" env-default:"147274"`
}

func MustLoad() *Config {
	godotenv.Load(".env")

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Environment configuration error: %v", err)
	}

	return &cfg
}
