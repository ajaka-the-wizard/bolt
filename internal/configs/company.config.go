package configs

import (
	"context"
	"log/slog"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/database"
	"github.com/ajaka-the-wizard/bolt/internal/models"
	"github.com/go-playground/validator/v10"
)

// MustLoadCompany loads the company details from the database. It'll panic should for any reason it couldn't load the data or validate that the required fields are present
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
