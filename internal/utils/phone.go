package utils

import (
	"errors"
	"strings"
)

// ValidateAndNormalizePhone standardizes mobile numbers.
// It trims whitespace, removes a leading "+91" or " 91", and ensures the final string is exactly 10 digits.
func ValidateAndNormalizePhone(phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	if strings.HasPrefix(phone, "+91") {
		phone = strings.TrimPrefix(phone, "+91")
	} else if strings.HasPrefix(phone, " 91") {
		phone = strings.TrimPrefix(phone, " 91")
	}
	
	// Remove any intermediate spaces
	phone = strings.ReplaceAll(phone, " ", "")

	if len(phone) != 10 {
		return "", errors.New("mobile number must be exactly 10 digits (without country code)")
	}
	return phone, nil
}
