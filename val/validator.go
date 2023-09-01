package val

import (
	"fmt"
	"net/mail"
	"regexp"

	"github.com/labasubagia/simplebank/util"
)

var (
	isValidUsername  = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
	isValidUFullName = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString
)

func ValidateString(value string, min, max int) error {
	n := len(value)
	if n < min || n > max {
		return fmt.Errorf("must contain from %d-%d characters", min, max)
	}
	return nil
}

func ValidateUsername(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}
	if !isValidUsername(value) {
		return fmt.Errorf("must contain only lowercase letters, digits or underscores")
	}
	return nil
}

func ValidateFullName(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}
	if !isValidUFullName(value) {
		return fmt.Errorf("must contain only letters or spaces")
	}
	return nil
}

func ValidatePassword(value string) error {
	return ValidateString(value, 8, 200)
}

func ValidateEmail(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("is not an valid email address")
	}
	return nil
}

func ValidateEmailID(value int64) error {
	if value <= 0 {
		return fmt.Errorf("must be a positive integer")
	}
	return nil
}

func ValidateID(value int64) error {
	if value <= 0 {
		return fmt.Errorf("id must be positive integer")
	}
	return nil
}

func ValidateCurrency(currency string) error {
	if !util.IsSupportedCurrency(currency) {
		return util.ErrInvalidCurrency
	}
	return nil
}

func ValidateSecretCode(value string) error {
	return ValidateString(value, 32, 128)
}
