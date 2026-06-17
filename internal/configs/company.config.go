package configs

import (
	"context"
	"log/slog"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/database"
	"github.com/ajaka-the-wizard/bolt/internal/models"
	"github.com/go-playground/validator/v10"
)

func MustLoadCompany(db *database.Repo, logger *slog.Logger, ctx context.Context) *models.CompanyInfo {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	company, err := db.LoadCompany(ctx)

	if err != nil {
		logger.Error("Couldnt load company", "error", err)
		panic(err)
	}

	v := validator.New()

	if err := v.Struct(company); err != nil {
		logger.Error("Missing company fields. ", "error", err.Error())
		panic(err)
	}
	logger.Info("Company details loaded successfully")

	return company
}
