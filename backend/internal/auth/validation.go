package auth

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func ValidateEmail(email string) bool {
	// Basic email validation logic.
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	// Check for the presence of '@' symbol in the email.
	if !strings.Contains(email, "@") {
		return false
	}
	return true
}

func ValidatePassword(password string) (bool, string) {
	if len(password) > 1_024 {
		return false, "password must not exceed 1024 bytes"
	}

	length := utf8.RuneCountInString(password)

	if length < 8 {
		return false, "password must be at least 8 characters long"
	}

	// Long passwords/passphrases need no composition rules.
	if length >= 15 {
		return true, ""
	}

	var hasUppercase bool
	var hasSymbol bool

	for _, character := range password {
		if unicode.IsUpper(character) {
			hasUppercase = true
		}

		if unicode.IsPunct(character) || unicode.IsSymbol(character) {
			hasSymbol = true
		}
	}

	if !hasUppercase || !hasSymbol {
		return false, "passwords shorter than 15 characters require an uppercase letter and a symbol"
	}

	return true, ""
}
