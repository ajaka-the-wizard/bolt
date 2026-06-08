package configs

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	PORT string
}

func LoadEnv(logger *slog.Logger) *Env {
	err := godotenv.Load()
	if err != nil {
		logger.Error("environment file not found", "error", err.Error())
		panic("Env file not found")
	}
	envs := Env{
		PORT: os.Getenv("PORT"),
	}
	return &envs
}
