package configs

import (
	"log/slog"

	"github.com/spf13/viper"
)

type Env struct {
	PORT           string
	DATABASE_URL   string
	REDIS_ADDR     string
	REDIS_PASSWORD string
	SHARED_SECRET  string
}

func LoadEnv(logger *slog.Logger) *Env {
	v := viper.New()
	v.SetConfigFile(".env")

	if err := v.ReadInConfig(); err != nil {
		logger.Error("Couldn't read env file", "error", err)
		panic(err)
	}

	var env Env
	if err := v.UnmarshalExact(&env); err != nil {
		logger.Error("Failed to map env config", "error", err)
		panic(err)
	}
	logger.Info("Env loaded successfully")
	return &env
}
