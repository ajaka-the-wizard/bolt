package configs

import (
	"log/slog"

	"github.com/ajaka-the-wizard/bolt/internal/domain"
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

func LoadEnvAndCompany(logger *slog.Logger) (*Env, *domain.CompanyInfo) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(".env")
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		logger.Error("Couldn't read env file, using system enviroments variables")
	}

	var env Env
	var company domain.CompanyInfo
	if err := v.Unmarshal(&env); err != nil {
		logger.Error("Failed to map env config", "error", err)
		panic(err)
	}
	validate := validator.New()
	if err := validate.Struct(&env); err != nil {
		logger.Error("Missing env fields. ", "error", err.Error())
		panic(err)
	}
	logger.Info("Env loaded successfully")

	if err := v.Unmarshal(&company); err != nil {
		logger.Error("Failed to map company config", "error", err.Error())
		panic(err)
	}
	if err := validate.Struct(&company); err != nil {
		logger.Error("Missing company fields. ", "error", err.Error())
		panic(err)
	}

	logger.Info("Company data loaded successfully")

	return &env, &company
}
