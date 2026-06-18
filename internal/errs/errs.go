package errs

import "errors"

var (
	ErrOrderNoExists   = errors.New("order with this id doesnt exists")
	ErrCompanyNoExists = errors.New("company details doesn't exist in the database")
)
