package auth

import (
	"strings"
	"testing"
)

func TestHashPassword_Success(t *testing.T) {
	password := "SecretPass123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	if !CheckPasswordHash(password, hash) {
		t.Fatal("expected CheckPasswordHash to return true for matching password")
	}

	if CheckPasswordHash("WrongPassword!", hash) {
		t.Fatal("expected CheckPasswordHash to return false for non-matching password")
	}
}

func TestHashPassword_LengthValidation(t *testing.T) {
	// Too short (< 8)
	_, err := HashPassword("short")
	if err != ErrPasswordTooShort {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}

	// Too long (> 72)
	longPass := strings.Repeat("a", 73)
	_, err = HashPassword(longPass)
	if err != ErrPasswordTooLong {
		t.Fatalf("expected ErrPasswordTooLong, got %v", err)
	}
}
