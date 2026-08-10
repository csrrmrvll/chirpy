package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestMakeJWT(t *testing.T) {
	// Test data
	userID := "123e4567-e89b-12d3-a456-426614174000"
	tokenSecret := "mysecretkey"
	expiresIn := 1 * time.Minute
	// Call the function
	token, err := MakeJWT(uuid.MustParse(userID), tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}
	if token == "" {
		t.Fatalf("MakeJWT() returned empty token")
	}
}

func TestValidateJWT(t *testing.T) {
	// Test data
	userID := "123e4567-e89b-12d3-a456-426614174000"
	tokenSecret := "mysecretkey"
	expiresIn := 1 * time.Minute

	// Create a valid token
	validToken, err := MakeJWT(uuid.MustParse(userID), tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	id, err := ValidateJWT(validToken, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}
	if id.String() != userID {
		t.Fatalf("ValidateJWT() returned wrong user ID: got %v, want %v", id.String(), userID)
	}
}
