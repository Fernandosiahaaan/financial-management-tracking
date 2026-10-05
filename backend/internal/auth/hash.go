package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var (
	ErrPasswordTooShort = errors.New("password must be at least 8 characters long")
	ErrPasswordTooLong  = errors.New("password cannot exceed 72 characters")
	ErrInvalidPIN       = errors.New("pin must be exactly 6 digits")
)

// HashPassword hashes a plaintext password using bcrypt with cost 12.
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", ErrPasswordTooShort
	}
	if len(password) > 72 {
		return "", ErrPasswordTooLong
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(bytes), nil
}

// HashPIN validates and hashes a 6-digit PIN using bcrypt.
func HashPIN(pin string) (string, error) {
	if len(pin) != 6 {
		return "", ErrInvalidPIN
	}
	for _, ch := range pin {
		if ch < '0' || ch > '9' {
			return "", ErrInvalidPIN
		}
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(pin), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash pin: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a plaintext password with a bcrypt hashed password.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

