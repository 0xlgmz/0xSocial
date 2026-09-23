package httpapi

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	// Test cases for validateEmail function
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"invalid-email", false},
	}
	for _, tt := range tests {
		actual := validateEmail(tt.email)
		if actual != tt.expected {
			t.Errorf("validateEmail(%q) = %v, want %v", tt.email, actual, tt.expected)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	// Test cases for validatePassword function
	tests := []struct {
		password string
		expected bool
	}{
		{"short", false},
		{"thispasswordiswaytoolongtobeacceptedbythesystemandshouldfailthevalidationcheck", false},
		{"validpassword1@", true},
		{"validpassword1#", true},
		{"validpassword1$", true},
		{"validpassword1%", true},
		{"invalidpassword", false}, // No special character
	}
	for _, tt := range tests {
		actual, _ := validatePassword(tt.password)
		if actual != tt.expected {
			t.Errorf("validatePassword(%q) = %v, want %v", tt.password, actual, tt.expected)
		}
	}
}
