package utils

import "errors"

// Error error list
var (
	ErrInvalidSymbol = errors.New("log: layout contains invalid symbol")

	ErrNumberOutOfRange = errors.New("number: out of range")
	ErrTimeWrongFormat  = errors.New("time: wrong format")

	ErrNotSupportAuth = errors.New("smtp: server doesn't support AUTH")
	ErrNoReciver      = errors.New("mail: no reciver")
	ErrNilConfig      = errors.New("mail: nil config")
)
