package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockRepo struct {
	users       map[string]User
	usersByID   map[uuid.UUID]User
	settings    map[uuid.UUID]UserSettings
	failCreate  error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users:     make(map[string]User),
		usersByID: make(map[uuid.UUID]User),
		settings:  make(map[uuid.UUID]UserSettings),
	}
}

func (m *mockRepo) CreateUser(ctx context.Context, user User, defaultCycleStartDay int) (*User, error) {
	if m.failCreate != nil {
		return nil, m.failCreate
	}
	if _, exists := m.users[user.Email]; exists {
		return nil, ErrUserAlreadyExists
	}

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	m.users[user.Email] = user
	m.usersByID[user.ID] = user
	m.settings[user.ID] = UserSettings{
		UserID:        user.ID,
		CycleStartDay: defaultCycleStartDay,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	return &user, nil
}

func (m *mockRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return &u, nil
}

func (m *mockRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return &u, nil
}

func (m *mockRepo) GetUserSettings(ctx context.Context, userID uuid.UUID) (*UserSettings, error) {
	s, ok := m.settings[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return &s, nil
}

func (m *mockRepo) UpdateUserSettings(ctx context.Context, userID uuid.UUID, cycleStartDay int) (*UserSettings, error) {
	s, ok := m.settings[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	s.CycleStartDay = cycleStartDay
	s.UpdatedAt = time.Now()
	m.settings[userID] = s
	return &s, nil
}

func (m *mockRepo) GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	u, ok := m.usersByID[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	s, ok := m.settings[userID]
	cycleDay := 1
	if ok {
		cycleDay = s.CycleStartDay
	}
	return &UserProfile{
		ID:            u.ID,
		Email:         u.Email,
		CycleStartDay: cycleDay,
		CreatedAt:     u.CreatedAt,
		HasPin:        u.PinHash != nil && *u.PinHash != "",
	}, nil
}

func (m *mockRepo) SetPinHash(ctx context.Context, userID uuid.UUID, pinHash *string) error {
	u, ok := m.usersByID[userID]
	if !ok {
		return ErrUserNotFound
	}
	u.PinHash = pinHash
	m.usersByID[userID] = u
	m.users[u.Email] = u
	return nil
}

func setupTestService() (*DefaultService, *mockRepo) {
	repo := newMockRepo()
	tokens := NewJWTService("test-secret-key-1234567890123456", 1*time.Hour)
	svc := NewService(repo, tokens)
	return svc, repo
}

func TestService_Register_Success(t *testing.T) {
	svc, _ := setupTestService()
	res, err := svc.Register(context.Background(), "user@example.com", "Password123!", 15)
	if err != nil {
		t.Fatalf("expected successful registration, got: %v", err)
	}

	if res.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if res.Profile.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", res.Profile.Email)
	}
	if res.Profile.CycleStartDay != 15 {
		t.Errorf("expected cycle start day 15, got %d", res.Profile.CycleStartDay)
	}
}

func TestService_Register_InvalidEmail(t *testing.T) {
	svc, _ := setupTestService()
	_, err := svc.Register(context.Background(), "invalid-email", "Password123!", 1)
	if err != ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestService_Register_WeakPassword(t *testing.T) {
	svc, _ := setupTestService()
	_, err := svc.Register(context.Background(), "user@example.com", "short", 1)
	if err != ErrPasswordTooShort {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestService_Register_DuplicateEmail(t *testing.T) {
	svc, _ := setupTestService()
	_, err := svc.Register(context.Background(), "user@example.com", "Password123!", 1)
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	_, err = svc.Register(context.Background(), "user@example.com", "Password123!", 1)
	if err != ErrUserAlreadyExists {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestService_Login_Success(t *testing.T) {
	svc, _ := setupTestService()
	_, err := svc.Register(context.Background(), "user@example.com", "Password123!", 25)
	if err != nil {
		t.Fatalf("setup registration failed: %v", err)
	}

	res, err := svc.Login(context.Background(), "user@example.com", "Password123!")
	if err != nil {
		t.Fatalf("expected successful login, got: %v", err)
	}

	if res.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if res.Profile.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", res.Profile.Email)
	}
	if res.Profile.CycleStartDay != 25 {
		t.Errorf("expected cycle start day 25, got %d", res.Profile.CycleStartDay)
	}
}

func TestService_Login_InvalidCredentials(t *testing.T) {
	svc, _ := setupTestService()
	_, _ = svc.Register(context.Background(), "user@example.com", "Password123!", 1)

	// Wrong password
	_, err := svc.Login(context.Background(), "user@example.com", "WrongPassword!")
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials for wrong password, got %v", err)
	}

	// Non-existent email
	_, err = svc.Login(context.Background(), "nobody@example.com", "Password123!")
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials for unknown user, got %v", err)
	}
}

func TestService_UpdateSettings(t *testing.T) {
	svc, _ := setupTestService()
	reg, _ := svc.Register(context.Background(), "user@example.com", "Password123!", 1)

	// Valid update
	updated, err := svc.UpdateSettings(context.Background(), reg.Profile.ID, 20)
	if err != nil {
		t.Fatalf("unexpected error updating settings: %v", err)
	}
	if updated.CycleStartDay != 20 {
		t.Errorf("expected cycle start day 20, got %d", updated.CycleStartDay)
	}

	// Invalid update (e.g. 32)
	_, err = svc.UpdateSettings(context.Background(), reg.Profile.ID, 32)
	if err != ErrInvalidCycleDay {
		t.Fatalf("expected ErrInvalidCycleDay, got %v", err)
	}
}

func TestService_PinAuth(t *testing.T) {
	svc, _ := setupTestService()
	reg, err := svc.Register(context.Background(), "pinuser@example.com", "Password123!", 1)
	if err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	// 1. PIN login before PIN is set should fail with ErrPinNotConfigured
	_, err = svc.PinLogin(context.Background(), "pinuser@example.com", "123456")
	if !errors.Is(err, ErrPinNotConfigured) {
		t.Fatalf("expected ErrPinNotConfigured, got %v", err)
	}

	// 2. Setting invalid PIN should fail
	err = svc.SetPin(context.Background(), reg.Profile.ID, "12345") // 5 digits
	if !errors.Is(err, ErrInvalidPIN) {
		t.Fatalf("expected ErrInvalidPIN for 5 digits, got %v", err)
	}

	err = svc.SetPin(context.Background(), reg.Profile.ID, "12345a") // non-digit
	if !errors.Is(err, ErrInvalidPIN) {
		t.Fatalf("expected ErrInvalidPIN for non-digits, got %v", err)
	}

	// 3. Setting valid PIN
	err = svc.SetPin(context.Background(), reg.Profile.ID, "123456")
	if err != nil {
		t.Fatalf("failed to set PIN: %v", err)
	}

	// 4. Verify profile shows HasPin = true
	prof, err := svc.GetProfile(context.Background(), reg.Profile.ID)
	if err != nil {
		t.Fatalf("failed to get profile: %v", err)
	}
	if !prof.HasPin {
		t.Fatalf("expected HasPin to be true")
	}

	// 5. Login with wrong PIN should fail
	_, err = svc.PinLogin(context.Background(), "pinuser@example.com", "654321")
	if !errors.Is(err, ErrInvalidPinCredentials) {
		t.Fatalf("expected ErrInvalidPinCredentials for wrong PIN, got %v", err)
	}

	// 6. Login with correct PIN should succeed and return token
	authRes, err := svc.PinLogin(context.Background(), "pinuser@example.com", "123456")
	if err != nil {
		t.Fatalf("unexpected error on pin login: %v", err)
	}
	if authRes.Token == "" {
		t.Fatalf("expected valid token from pin login")
	}
	if authRes.Profile.Email != "pinuser@example.com" {
		t.Errorf("expected email pinuser@example.com, got %s", authRes.Profile.Email)
	}

	// 7. Remove PIN
	err = svc.RemovePin(context.Background(), reg.Profile.ID)
	if err != nil {
		t.Fatalf("failed to remove PIN: %v", err)
	}

	// 8. Profile should have HasPin = false
	prof, err = svc.GetProfile(context.Background(), reg.Profile.ID)
	if err != nil {
		t.Fatalf("failed to get profile: %v", err)
	}
	if prof.HasPin {
		t.Fatalf("expected HasPin to be false after removal")
	}

	// 9. Login with PIN after removal should fail with ErrPinNotConfigured
	_, err = svc.PinLogin(context.Background(), "pinuser@example.com", "123456")
	if !errors.Is(err, ErrPinNotConfigured) {
		t.Fatalf("expected ErrPinNotConfigured after removal, got %v", err)
	}
}

