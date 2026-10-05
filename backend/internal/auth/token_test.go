package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTService_GenerateAndValidate(t *testing.T) {
	svc := NewJWTService("test-secret-key-1234567890123456", 1*time.Hour)
	user := User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	tokenStr, err := svc.GenerateToken(user)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	claims, err := svc.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, claims.UserID)
	}
	if claims.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, claims.Email)
	}
}

func TestJWTService_InvalidSecret(t *testing.T) {
	svc1 := NewJWTService("secret-one-12345678901234567890", 1*time.Hour)
	svc2 := NewJWTService("secret-two-12345678901234567890", 1*time.Hour)

	user := User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	tokenStr, _ := svc1.GenerateToken(user)
	_, err := svc2.ValidateToken(tokenStr)
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken when validating with different secret, got %v", err)
	}
}

func TestJWTService_ExpiredToken(t *testing.T) {
	// Expired immediately (-1 second)
	svc := NewJWTService("test-secret-key-1234567890123456", -1*time.Second)
	user := User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	tokenStr, _ := svc.GenerateToken(user)
	_, err := svc.ValidateToken(tokenStr)
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for expired token, got %v", err)
	}
}
