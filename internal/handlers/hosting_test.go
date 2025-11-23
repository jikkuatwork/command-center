package handlers

import (
	"testing"
)

func TestValidateEnvVarName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		// Valid names
		{"simple uppercase", "API_KEY", false, ""},
		{"single letter", "A", false, ""},
		{"with numbers", "API_KEY_V2", false, ""},
		{"all caps no underscore", "APIKEY", false, ""},
		{"long valid name", "MY_SUPER_LONG_ENVIRONMENT_VARIABLE_NAME", false, ""},

		// Invalid - empty
		{"empty string", "", true, "cannot be empty"},
		{"whitespace only", "   ", true, "cannot be empty"},

		// Invalid - format
		{"lowercase", "api_key", true, "uppercase"},
		{"starts with number", "1API_KEY", true, "starting with a letter"},
		{"starts with underscore", "_API_KEY", true, "starting with a letter"},
		{"contains hyphen", "API-KEY", true, "uppercase"},
		{"contains space", "API KEY", true, "uppercase"},
		{"mixed case", "Api_Key", true, "uppercase"},

		// Invalid - too long
		{"too long", string(make([]byte, 129)), true, "too long"},

		// Dangerous system vars
		{"PATH blocked", "PATH", true, "reserved system"},
		{"LD_PRELOAD blocked", "LD_PRELOAD", true, "reserved system"},
		{"LD_LIBRARY_PATH blocked", "LD_LIBRARY_PATH", true, "reserved system"},
		{"HOME blocked", "HOME", true, "reserved system"},
		{"USER blocked", "USER", true, "reserved system"},
		{"SHELL blocked", "SHELL", true, "reserved system"},
		{"NODE_OPTIONS blocked", "NODE_OPTIONS", true, "reserved system"},
		{"HTTP_PROXY blocked", "HTTP_PROXY", true, "reserved system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEnvVarName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateEnvVarName(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("validateEnvVarName(%q) unexpected error: %v", tt.input, err)
				}
			}
		})
	}
}

func TestValidEnvVarNameRegex(t *testing.T) {
	// Test the regex directly
	validNames := []string{
		"A", "ABC", "A1", "A_B", "ABC_DEF_123", "X1Y2Z3",
	}

	for _, name := range validNames {
		if !validEnvVarName.MatchString(name) {
			t.Errorf("Regex should match valid name: %s", name)
		}
	}

	invalidNames := []string{
		"", "a", "1A", "_A", "A-B", "A B", "a_b", "Ab",
	}

	for _, name := range invalidNames {
		if validEnvVarName.MatchString(name) {
			t.Errorf("Regex should not match invalid name: %s", name)
		}
	}
}

func TestDangerousEnvVars(t *testing.T) {
	// Ensure all dangerous vars are in the map
	expected := []string{
		"PATH", "LD_PRELOAD", "LD_LIBRARY_PATH",
		"HOME", "USER", "SHELL", "PWD",
		"TERM", "LANG", "LC_ALL",
		"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY",
		"NODE_OPTIONS", "NODE_PATH",
	}

	for _, name := range expected {
		if !dangerousEnvVars[name] {
			t.Errorf("dangerousEnvVars should contain %s", name)
		}
	}
}

func TestValidationError(t *testing.T) {
	err := &validationError{"test message"}
	if err.Error() != "test message" {
		t.Errorf("Error() = %s, want 'test message'", err.Error())
	}
}

func TestValidateEnvVarNameEdgeCases(t *testing.T) {
	// Boundary: exactly 128 chars (should pass)
	longValid := make([]byte, 128)
	for i := range longValid {
		if i == 0 {
			longValid[i] = 'A'
		} else {
			longValid[i] = 'B'
		}
	}
	if err := validateEnvVarName(string(longValid)); err != nil {
		t.Errorf("128-char name should be valid: %v", err)
	}

	// Boundary: 129 chars (should fail)
	tooLong := make([]byte, 129)
	for i := range tooLong {
		if i == 0 {
			tooLong[i] = 'A'
		} else {
			tooLong[i] = 'B'
		}
	}
	if err := validateEnvVarName(string(tooLong)); err == nil {
		t.Error("129-char name should be invalid")
	}
}
