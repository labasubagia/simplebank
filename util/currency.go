package util

import "errors"

const (
	USD = "USD"
	EUR = "EUR"
	IDR = "IDR"
)

var ErrInvalidCurrency = errors.New("currency invalid")

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, IDR:
		return true
	}
	return false
}
