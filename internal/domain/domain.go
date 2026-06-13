package domain

// Add other detail as needed
type CompanyInfo struct {
	Name         string `mapstructure:"NAME" validate:"required"`
	AddressLine1 string `mapstructure:"ADDRESS1" validate:"required"`
	AddressLine2 string `mapstructure:"ADDRESS2"`
	City         string `mapstructure:"CITY" validate:"required"`
	State        string `mapstructure:"STATE" validate:"required"`
	PostalCode   string `mapstructure:"POSTALCODE" validate:"required"`
	Country      string `mapstructure:"COUNTRY" validate:"required"`
	Phone        string `mapstructure:"PHONE" validate:"required"`
	Email        string `mapstructure:"EMAIL" validate:"required"`
	Website      string `mapstructure:"WEBSITE" validate:"required"`
	TaxID        string `mapstructure:"TAXID" validate:"required"`
}
