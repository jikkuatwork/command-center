package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost is the cost factor for bcrypt hashing (higher = more secure but slower)
	BcryptCost = 12

	// MinPasswordLength is the minimum allowed password length
	MinPasswordLength = 8
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	if len(password) < MinPasswordLength {
		return "", fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword compares a password with a hash
func VerifyPassword(password, hash string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	if hash == "" {
		return errors.New("hash cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("invalid password")
		}
		return fmt.Errorf("failed to verify password: %w", err)
	}

	return nil
}

// ValidatePasswordStrength checks password strength and returns recommendations
func ValidatePasswordStrength(password string) (isStrong bool, warnings []string) {
	if len(password) < MinPasswordLength {
		warnings = append(warnings, fmt.Sprintf("Password should be at least %d characters", MinPasswordLength))
		return false, warnings
	}

	// Check for basic strength criteria
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= '!' && char <= '/' || char >= ':' && char <= '@' || char >= '[' && char <= '`' || char >= '{' && char <= '~':
			hasSpecial = true
		}
	}

	// Add warnings for missing character types
	if !hasLower {
		warnings = append(warnings, "Consider adding lowercase letters")
	}
	if !hasUpper {
		warnings = append(warnings, "Consider adding uppercase letters")
	}
	if !hasDigit {
		warnings = append(warnings, "Consider adding numbers")
	}
	if !hasSpecial {
		warnings = append(warnings, "Consider adding special characters (!@#$%^&*)")
	}

	// Consider strong if it has at least 3 out of 4 criteria
	criteriaCount := 0
	if hasLower {
		criteriaCount++
	}
	if hasUpper {
		criteriaCount++
	}
	if hasDigit {
		criteriaCount++
	}
	if hasSpecial {
		criteriaCount++
	}

	isStrong = criteriaCount >= 3 && len(password) >= 12

	if len(password) < 12 {
		warnings = append(warnings, "For better security, use at least 12 characters")
	}

	return isStrong, warnings
}
