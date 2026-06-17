package domain

import "errors"

var (
	ErrOrderNoExists = errors.New("order with this id doesnt exists")
)
