package models

type Address struct {
	AddressLine1 string `db:"address_line1" json:"address_line1"`
	AddressLine2 string `db:"address_line2" json:"address_line2"`
	City         string `db:"city" json:"city"`
	State        string `db:"state" json:"state"`
	PostalCode   string `db:"postal_code" json:"postal_code"`
	Country      string `db:"country" json:"country"`
}
