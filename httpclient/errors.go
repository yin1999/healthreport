package httpclient

import (
	"errors"
)

// Error error list
var (
	ErrCannotParseData = errors.New("data: parse error")
	ErrInvalidSymbol   = errors.New("data: invalid symbol")
)
