package configs

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Env struct {
	PORT           string `mapstructure:"PORT" validate:"required"`
	DATABASE_URL   string `mapstructure:"DATABASE_URL" validate:"required"`
	REDIS_ADDR     string `mapstructure:"REDIS_ADDR" validate:"required"`
	REDIS_PASSWORD string `mapstructure:"REDIS_PASSWORD"`
	SHARED_SECRET  string `mapstructure:"SHARED_SECRET" validate:"required"`
	PRODUCTION     bool   `mapstructure:"PRODUCTION"`
}

// MustLoadEnv is responsible for loading and validating all environments variable used throughout the app. It will panic should any error occur during that process.
func MustLoadEnv(logger *slog.Logger) *Env {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(".env")
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		logger.Error("Couldn't read env file, using system enviroments variables")
	}

	var env Env
	if err := v.UnmarshalExact(&env); err != nil {
		logger.Error("Failed to map env config", "error", err)
		panic(err)
	}
	validate := validator.New()
	if err := validate.Struct(&env); err != nil {
		logger.Error("Missing env fields. ", "error", err.Error())
		panic(err)
	}
	logger.Info("Env loaded successfully")

	return &env
}
