package utils

import (
	"errors"
	"strings"
)

// ValidateAndNormalizePhone standardizes mobile numbers.
// It trims whitespace, removes a leading "+91" or " 91", and ensures the final string is exactly 10 digits.
func ValidateAndNormalizePhone(phone string) (string, error) {
	// Remove any spaces and '+' symbols
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "+", "")

	// If it has 12 digits and starts with 91, strip the 91
	if len(phone) == 12 && strings.HasPrefix(phone, "91") {
		phone = strings.TrimPrefix(phone, "91")
	}

	if len(phone) != 10 {
		return "", errors.New("mobile number must be exactly 10 digits (without country code)")
	}
	return phone, nil
}
