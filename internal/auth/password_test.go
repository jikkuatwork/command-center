package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	// Valid password
	hash, err := HashPassword("SecurePass123!")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Error("Hash should not be empty")
	}

	// Empty password
	_, err = HashPassword("")
	if err == nil {
		t.Error("HashPassword should fail for empty password")
	}

	// Short password
	_, err = HashPassword("short")
	if err == nil {
		t.Error("HashPassword should fail for short password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "SecurePass123!"
	hash, _ := HashPassword(password)

	// Correct password
	err := VerifyPassword(password, hash)
	if err != nil {
		t.Errorf("VerifyPassword failed for correct password: %v", err)
	}

	// Wrong password
	err = VerifyPassword("WrongPass", hash)
	if err == nil {
		t.Error("VerifyPassword should fail for wrong password")
	}

	// Empty password
	err = VerifyPassword("", hash)
	if err == nil {
		t.Error("VerifyPassword should fail for empty password")
	}

	// Empty hash
	err = VerifyPassword(password, "")
	if err == nil {
		t.Error("VerifyPassword should fail for empty hash")
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		wantStrong bool
	}{
		{"strong password", "SecurePass123!@#", true},
		{"weak - too short", "Abc1!", false},
		{"strong - 3 criteria met (no special)", "SecurePass123", true},      // lower, upper, digit = 3 criteria, 13 chars
		{"strong - 3 criteria met (no digits)", "SecurePassword!", true},     // lower, upper, special = 3 criteria, 15 chars
		{"weak - lowercase only", "securepassword", false},                   // only 1 criteria
		{"weak - short with 3 criteria", "SecPass1!", false},                 // 3 criteria but only 9 chars
		{"weak - 2 criteria only", "SECUREPASSWORD", false},                  // upper only, 14 chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strong, warnings := ValidatePasswordStrength(tt.password)
			if strong != tt.wantStrong {
				t.Errorf("ValidatePasswordStrength(%q) strong = %v, want %v (warnings: %v)",
					tt.password, strong, tt.wantStrong, warnings)
			}
		})
	}
}

func TestPasswordRoundTrip(t *testing.T) {
	password := "MySecurePassword123!"

	// Hash
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Different hash each time (bcrypt uses random salt)
	hash2, _ := HashPassword(password)
	if hash == hash2 {
		t.Error("Hashes should be different (bcrypt uses random salt)")
	}

	// But both should verify
	if err := VerifyPassword(password, hash); err != nil {
		t.Error("First hash should verify")
	}
	if err := VerifyPassword(password, hash2); err != nil {
		t.Error("Second hash should verify")
	}
}
