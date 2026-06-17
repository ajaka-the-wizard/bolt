package models

import "time"

type CompanyInfo struct {
	Name         string    `db:"name" validate:"required"`
	AddressLine1 string    `db:"address1" validate:"required"`
	AddressLine2 string    `db:"address2"`
	City         string    `db:"city" validate:"required"`
	State        string    `db:"state" validate:"required"`
	PostalCode   string    `db:"postal_code" validate:"required"`
	Country      string    `db:"country" validate:"required"`
	Phone        string    `db:"phone" validate:"required"`
	Email        string    `db:"email" validate:"required"`
	Website      string    `db:"website" validate:"required"`
	TaxID        string    `db:"tax_id" validate:"required"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
