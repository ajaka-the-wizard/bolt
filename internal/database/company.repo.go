package database

import (
	"context"
	"errors"

	"github.com/ajaka-the-wizard/bolt/internal/errs"
	"github.com/ajaka-the-wizard/bolt/internal/models"
	"github.com/jackc/pgx/v5"
)

// LoadCompany loads the company data from the database, returns it or any potential error.
func (r *Repo) LoadCompany(ctx context.Context) (*models.CompanyInfo, error) {
	query := `
	SELECT * from company
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	company, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[models.CompanyInfo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrCompanyNoExists
		}
		return nil, err
	}
	return &company, nil
}
